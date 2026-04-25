package cli

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/thismarioperez/squrl/assets"
	"github.com/thismarioperez/squrl/internal/scanner"
)

type appState int

const (
	stateIdle      appState = iota
	stateCountdown
	stateScanning
	stateResults
)

type scanResultMsg struct {
	results []string
	err     error
}

type keyMap struct {
	Scan  key.Binding
	Copy  key.Binding
	Clear key.Binding
	Quit  key.Binding
	Help  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Scan, k.Copy, k.Clear, k.Quit, k.Help}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Scan, k.Copy},
		{k.Clear, k.Quit},
		{k.Help},
	}
}

var defaultKeys = keyMap{
	Scan:  key.NewBinding(key.WithKeys("space", "r"), key.WithHelp("space/r", "scan")),
	Copy:  key.NewBinding(key.WithKeys("enter", "l"), key.WithHelp("enter/l", "copy")),
	Clear: key.NewBinding(key.WithKeys("h", "c"), key.WithHelp("h/c", "back")),
	Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:  key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
}

type resultItem struct{ value string }

func (r resultItem) Title() string       { return r.value }
func (r resultItem) Description() string { return "" }
func (r resultItem) FilterValue() string { return r.value }

type model struct {
	opts        ScanOptions
	state       appState
	timer       timer.Model
	list        list.Model
	help        help.Model
	keys        keyMap
	err         error
	exitCode    int
	ctx         context.Context
	banner      string
	bannerLines int
	width       int
	height      int
}

func initialModel(ctx context.Context, opts ScanOptions) model {
	banner := string(assets.CLIIcon())
	banner = strings.ReplaceAll(banner, "\x1b[?25l", "")
	banner = strings.ReplaceAll(banner, "\x1b[?25h", "")

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New(nil, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetStatusBarItemName("QR code", "QR codes")

	m := model{
		opts:        opts,
		state:       stateIdle,
		ctx:         ctx,
		banner:      banner,
		bannerLines: strings.Count(banner, "\n") + 1,
		list:        l,
		help:        help.New(),
		keys:        defaultKeys,
	}
	if opts.Delay == 0 {
		m.state = stateScanning
	} else {
		m.state = stateCountdown
		m.timer = timer.New(time.Duration(opts.Delay) * time.Second)
	}
	return m
}

func (m model) Init() tea.Cmd {
	if m.state == stateScanning {
		return tea.Batch(doScan(m.ctx), waitForCtx(m.ctx))
	}
	return tea.Batch(m.timer.Init(), waitForCtx(m.ctx))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		listHeight := max(3, msg.Height-m.bannerLines-5)
		m.list.SetSize(msg.Width, listHeight)
		m.help.SetWidth(msg.Width)

	case tea.KeyPressMsg:
		switch msg.String() {
		case "?":
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case "q", "ctrl+c":
			return m, tea.Quit
		case "h", "c":
			m.state = stateIdle
			m.err = nil
			m.list.SetItems(nil)
			return m, nil
		case "space", "r":
			if m.state == stateIdle || m.state == stateResults {
				m.err = nil
				m.list.SetItems(nil)
				return m, m.startScanSequence()
			}
		case "enter", "l":
			if m.state == stateResults {
				if item, ok := m.list.SelectedItem().(resultItem); ok {
					if err := clipboard.WriteAll(item.value); err != nil {
						slog.Error("clipboard write failed", "err", err)
					}
				}
				return m, nil
			}
		}
		if m.state == stateResults {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd

	case timer.TimeoutMsg:
		m.state = stateScanning
		return m, doScan(m.ctx)

	case scanResultMsg:
		m.state = stateResults
		m.err = msg.err
		items := make([]list.Item, len(msg.results))
		for i, r := range msg.results {
			items[i] = resultItem{r}
		}
		cmd := m.list.SetItems(items)
		m.list.Title = fmt.Sprintf("%d QR codes found", len(msg.results))
		if msg.err != nil {
			m.exitCode = 2
		} else if len(msg.results) == 0 {
			m.exitCode = 1
		} else {
			m.exitCode = 0
		}
		return m, cmd

	case tea.QuitMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString(m.banner)
	b.WriteString(m.help.View(m.keys))
	b.WriteString("\n")
	b.WriteString("\n")

	switch m.state {
	case stateIdle:
		b.WriteString("Ready to scan.\n")
	case stateCountdown:
		fmt.Fprintf(&b, "Scanning in %s...\n", m.timer.View())
	case stateScanning:
		b.WriteString("Scanning...\n")
	case stateResults:
		if m.err != nil {
			if errors.Is(m.err, context.Canceled) {
				b.WriteString("Cancelled.\n")
			} else {
				b.WriteString("Error: " + m.err.Error() + "\n")
			}
		} else {
			b.WriteString(m.list.View())
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func (m *model) startScanSequence() tea.Cmd {
	if m.opts.Delay == 0 {
		m.state = stateScanning
		return doScan(m.ctx)
	}
	m.state = stateCountdown
	m.timer = timer.New(time.Duration(m.opts.Delay) * time.Second)
	return m.timer.Init()
}

func doScan(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		results, err := scanner.ScanAllScreens(ctx)
		return scanResultMsg{results: results, err: err}
	}
}

func waitForCtx(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		return tea.QuitMsg{}
	}
}
