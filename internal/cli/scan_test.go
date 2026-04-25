package cli

import (
	"testing"
)

func TestParseScanArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    ScanOptions
		wantErr bool
	}{
		{
			name: "defaults",
			args: []string{},
			want: ScanOptions{Delay: 3, NonInteractive: false},
		},
		{
			name: "-n sets NonInteractive",
			args: []string{"-n"},
			want: ScanOptions{Delay: 3, NonInteractive: true},
		},
		{
			name: "--non-interactive sets NonInteractive",
			args: []string{"--non-interactive"},
			want: ScanOptions{Delay: 3, NonInteractive: true},
		},
		{
			name: "--delay overrides default",
			args: []string{"--delay", "5"},
			want: ScanOptions{Delay: 5, NonInteractive: false},
		},
		{
			name: "-D shorthand for delay",
			args: []string{"-D", "0"},
			want: ScanOptions{Delay: 0, NonInteractive: false},
		},
		{
			name:    "negative delay returns error",
			args:    []string{"--delay", "-1"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseScanArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseScanArgs(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseScanArgs(%v) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}
