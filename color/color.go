package color

//go:generate stringify Parameter

import (
	"fmt"
	"strconv"
	"strings"
)

// Parameter defines an SGR parameter
// See: https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
type Parameter uint8

const Escape = "\x1b"

// Base params.
const (
	Reset Parameter = iota
	Bold
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkRapid
	ReverseVideo
	Conceal
	CrossedOut
	PrimaryFont
)

// Foreground.
const (
	FgBlack Parameter = iota + 30
	FgRed
	FgGreen
	FgYellow
	FgBlue
	FgMagenta
	FgCyan
	FgWhite
)

// Foreground Bright.
const (
	FgBBlack Parameter = iota + 90
	FgBRed
	FgBGreen
	FgBYellow
	FgBBlue
	FgBMagenta
	FgBCyan
	FgBWhite
)

// Background.
const (
	BgBlack Parameter = iota + 40
	BgRed
	BgGreen
	BgYellow
	BgBlue
	BgMagenta
	BgCyan
	BgWhite
)

// Background Bright.
const (
	BgBBlack Parameter = iota + 100
	BgBRed
	BgBGreen
	BgBYellow
	BgBBlue
	BgBMagenta
	BgBCyan
	BgBWhite
)

func colorSequence(colors ...Parameter) string {
	format := make([]string, len(colors))
	for i, v := range colors {
		format[i] = strconv.Itoa(int(v))
	}

	return strings.Join(format, ";")
}

func color(colors ...Parameter) string {
	return fmt.Sprintf("%s[%sm", Escape, colorSequence(colors...))
}

func ByNames(name ...string) string {
	if len(name) == 0 {
		return ""
	}

	if len(name) == 1 {
		return ByName(name[0])
	}

	params := make([]Parameter, len(name))

	for i, n := range name {
		params[i] = NewParameter(n)
	}

	return color(params...)
}

func ByName(name string) string {
	return color(NewParameter(name))
}

func Clear() string {
	return fmt.Sprintf("%s[%dm", Escape, Reset)
}

func Red() string {
	return color(FgRed)
}

func Green() string {
	return color(FgGreen)
}

func Blue() string {
	return color(FgBlue)
}

func Yellow() string {
	return color(FgYellow)
}

func Magenta() string {
	return color(FgMagenta)
}

func Cyan() string {
	return color(FgCyan)
}
