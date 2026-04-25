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
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/timer"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	stateCopied
)

// listHeightOffset accounts for the top newline, separator line,
// blank line after separator, blank line after content, and bottom padding.
const listHeightOffset = 5

type scanResultMsg struct {
	results []string
	err     error
}

type copiedDismissMsg struct{}

type keyMap struct {
	Scan  key.Binding
	Clear key.Binding
	Quit  key.Binding
	Help  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Scan, k.Clear},
		{k.Quit, k.Help},
	}
}

type listKeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	PrevPage   key.Binding
	NextPage   key.Binding
	Copy       key.Binding
}

func (k listKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.CursorUp, k.CursorDown, k.Copy}
}
func (k listKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.CursorUp, k.CursorDown, k.PrevPage, k.NextPage},
		{k.Copy},
	}
}

var contentStyle = lipgloss.NewStyle().PaddingLeft(2)

var defaultKeys = keyMap{
	Scan:  key.NewBinding(key.WithKeys("space", "r"), key.WithHelp("space/r", "scan")),
	Clear: key.NewBinding(key.WithKeys("esc", "c"), key.WithHelp("esc/c", "clear")),
	Quit:  key.NewBinding(key.WithKeys("ctrl+c","q"), key.WithHelp("ctrl+c/q", "quit")),
	Help:  key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
}

var defaultListKeys = listKeyMap{
	CursorUp:   key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	CursorDown: key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	PrevPage:   key.NewBinding(key.WithKeys("left", "h", "pgup"), key.WithHelp("←/h/pgup", "prev page")),
	NextPage:   key.NewBinding(key.WithKeys("right", "l", "pgdown"), key.WithHelp("→/l/pgdn", "next page")),
	Copy:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "copy")),
}

type resultItem struct{ value string }

func (r resultItem) Title() string       { return r.value }
func (r resultItem) Description() string { return "" }
func (r resultItem) FilterValue() string { return r.value }

type model struct {
	opts           ScanOptions
	version        string
	state          appState
	spinner        spinner.Model
	timer          timer.Model
	list           list.Model
	help           help.Model
	listHelp       help.Model
	keys           keyMap
	listKeys       listKeyMap
	err            error
	exitCode       int
	copiedValue    string
	ctx            context.Context
	renderedBanner string
	bannerLines    int
	bannerWidth    int
	helpWidth      int
	width          int
	height         int
}

func initialModel(ctx context.Context, opts ScanOptions, version string) model {
	banner := string(assets.CLIIcon())
	banner = strings.ReplaceAll(banner, "\x1b[?25l", "")
	banner = strings.ReplaceAll(banner, "\x1b[?25h", "")
	banner = strings.TrimRight(banner, "\n")

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	l := list.New(nil, delegate, 0, 0)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	l.SetStatusBarItemName("QR code", "QR codes")

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#c96b3a"))

	renderedBanner := contentStyle.Render(banner)
	m := model{
		opts:           opts,
		version:        version,
		state:          stateIdle,
		spinner:        sp,
		ctx:            ctx,
		renderedBanner: renderedBanner,
		bannerLines:    strings.Count(banner, "\n") + 1,
		bannerWidth:    lipgloss.Width(renderedBanner),
		list:           l,
		help:           help.New(),
		listHelp:       help.New(),
		keys:           defaultKeys,
		listKeys:       defaultListKeys,
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
		return tea.Batch(doScan(m.ctx), waitForCtx(m.ctx), m.spinner.Tick)
	}
	return tea.Batch(m.timer.Init(), waitForCtx(m.ctx), m.spinner.Tick)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		listHeight := max(3, msg.Height-m.bannerLines-listHeightOffset)
		m.list.SetSize(msg.Width, listHeight)
		m.helpWidth = max(0, msg.Width-m.bannerWidth)
		m.help.SetWidth(m.helpWidth)
		m.listHelp.SetWidth(msg.Width)

	case copiedDismissMsg:
		if m.state == stateCopied {
			m.state = stateResults
		}
		return m, nil

	case tea.KeyPressMsg:
		if m.state == stateCopied {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			default:
				m.state = stateResults
				return m, nil
			}
		}
		switch msg.String() {
		case "?":
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc", "c":
			m.state = stateIdle
			m.err = nil
			m.list.SetItems(nil)
			return m, nil
		case "space", "r":
			if m.state == stateIdle || m.state == stateResults {
				m.err = nil
				m.list.SetItems(nil)
				if m.opts.Delay == 0 {
					m.state = stateScanning
					return m, tea.Batch(doScan(m.ctx), m.spinner.Tick)
				}
				m.state = stateCountdown
				m.timer = timer.New(time.Duration(m.opts.Delay) * time.Second)
				return m, tea.Batch(m.timer.Init(), m.spinner.Tick)
			}
		case "enter":
			if m.state == stateResults {
				if item, ok := m.list.SelectedItem().(resultItem); ok {
					if err := clipboard.WriteAll(item.value); err != nil {
						slog.Error("clipboard write failed", "err", err)
					} else {
						m.copiedValue = item.value
						m.state = stateCopied
						return m, scheduleDismiss()
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

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

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
		m.list.Title = fmt.Sprintf("%d QR codes found", len(msg.results))
		cmd := m.list.SetItems(items)
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
	b.WriteString("\n")
	title := "squrl"
	version := m.help.Styles.ShortKey.Render("v" + m.version)
	rightPanel := title + " " + version + "\n\n" + m.help.View(m.keys)
	header := lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.renderedBanner,
		lipgloss.NewStyle().PaddingLeft(2).Render(rightPanel),
	)
	b.WriteString(header)
	b.WriteString("\n")
	ruleStyle := lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2)
	b.WriteString(ruleStyle.Render(strings.Repeat("─", max(0, m.width-4))) + "\n")
	b.WriteString("\n")

	switch m.state {
	case stateIdle:
		b.WriteString(contentStyle.Render("Ready to scan.") + "\n")
	case stateCountdown:
		b.WriteString(contentStyle.Render(m.spinner.View()+" Scanning in "+m.timer.View()+"...") + "\n")
	case stateScanning:
		b.WriteString(contentStyle.Render(m.spinner.View()+" Scanning...") + "\n")
	case stateCopied:
		truncated := m.copiedValue
		maxLen := max(0, m.width-20)
		if len(truncated) > maxLen {
			truncated = truncated[:maxLen] + "…"
		}
		b.WriteString(contentStyle.Render(
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("2")).
				Bold(true).
				Render("✓ Copied: "+truncated),
		) + "\n")
	case stateResults:
		if m.err != nil {
			if errors.Is(m.err, context.Canceled) {
				b.WriteString(contentStyle.Render("Cancelled.") + "\n")
			} else {
				b.WriteString(contentStyle.Render("Error: "+m.err.Error()) + "\n")
			}
		} else if len(m.list.Items()) == 0 {
			b.WriteString(contentStyle.Render("No QR codes found.") + "\n")
		} else {
			b.WriteString(contentStyle.Render(m.list.View()))
			b.WriteString("\n")
			b.WriteString(contentStyle.Render(m.listHelp.View(m.listKeys)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

func scheduleDismiss() tea.Cmd {
	return tea.Tick(1500*time.Millisecond, func(time.Time) tea.Msg {
		return copiedDismissMsg{}
	})
}

func doScan(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		results, err := scanner.ScanAllScreens(ctx)
		return scanResultMsg{results: results, err: err}
	}
}

// waitForCtx bridges context cancellation into the Bubble Tea event loop.
// The spawned goroutine blocks until ctx is done; since stop() is deferred
// in main immediately after p.Run() returns, the window is minimal.
func waitForCtx(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		return tea.QuitMsg{}
	}
}
