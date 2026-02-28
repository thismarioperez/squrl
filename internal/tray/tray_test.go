package tray

import "testing"

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