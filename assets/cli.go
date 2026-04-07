package assets

import _ "embed"

//go:embed cli_ansi.txt
var cliIcon []byte

// CLIIcon returns the pre-rendered ANSI art for the CLI banner.
func CLIIcon() []byte { return cliIcon }
