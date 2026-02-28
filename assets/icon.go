package assets

import _ "embed"

//go:embed menubar_22.png
var menubarIcon []byte

//go:embed menubar_44.png
var menubarIcon2x []byte

// Icon returns the 22×22 menu bar template icon PNG (1x).
func Icon() []byte { return menubarIcon }

// Icon2x returns the 44×44 menu bar template icon PNG (2x / Retina).
func Icon2x() []byte { return menubarIcon2x }
