package tray

import (
	"errors"
	"testing"
)

func TestParseDarkMode(t *testing.T) {
	tests := []struct {
		name string
		out  []byte
		err  error
		want bool
	}{
		{"gsettings error assumes dark", nil, errors.New("not found"), true},
		{"prefer-dark is dark", []byte("'prefer-dark'\n"), nil, true},
		{"prefer-light is light", []byte("'prefer-light'\n"), nil, false},
		{"prefer-none is dark", []byte("'prefer-none'\n"), nil, true},
		{"default is dark", []byte("'default'\n"), nil, true},
		{"empty output is dark", []byte(""), nil, true},
		{"case-insensitive prefer-light", []byte("Prefer-Light\n"), nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDarkMode(tt.out, tt.err)
			if got != tt.want {
				t.Errorf("parseDarkMode(%q, %v) = %v; want %v", tt.out, tt.err, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		n     int
		want  string
	}{
		{"empty string", "", 10, ""},
		{"shorter than limit", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"over limit", "hello world", 8, "hello w…"},
		{"unicode multibyte", "日本語テスト", 4, "日本語…"},
		{"limit of 1", "hi", 1, "…"},
		{"single rune at limit", "x", 1, "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.n)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q; want %q", tt.input, tt.n, got, tt.want)
			}
		})
	}
}