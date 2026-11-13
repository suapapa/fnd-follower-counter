// Package display formats follower counts and renders status on a TM1638 8-digit FND module.
package display

import (
	"fmt"

	"github.com/suapapa/go_devices/tm1638"
)

const sideDigitCount = 4

// Device represents the subset of a TM1638 module used for rendering.
type Device interface {
	SetChar(pos int, c rune, dot bool)
	SetLed(pos int, led tm1638.Color)
	SetString(str string)
}

// SideState holds the latest polled follower count and error state for one FND side.
type SideState struct {
	Count    int
	HasValue bool
	Err      error
}

// Render updates both the left and right 4-digit sides of dev.
func Render(dev Device, left, right SideState) {
	renderSide(dev, 0, left)
	renderSide(dev, sideDigitCount, right)
	if flusher, ok := dev.(interface{ Flush() }); ok {
		flusher.Flush()
	}
}

func renderSide(dev Device, offset int, state SideState) {
	chars := [sideDigitCount]rune{' ', 'E', 'r', 'r'}
	var dots [sideDigitCount]bool
	if state.HasValue {
		chars, dots = formatCount4(state.Count)
	}

	// Keep the middle dot lit on digit 3 to visually separate left and right sides
	// unless that digit already carries its own formatting.
	if offset == 0 && !dots[3] {
		dots[3] = true
	}

	ledColor := tm1638.Green
	if state.Err != nil {
		ledColor = tm1638.Red
	}

	for i := range sideDigitCount {
		dev.SetChar(offset+i, chars[i], dots[i])
		dev.SetLed(offset+i, ledColor)
	}
}

// formatCount4 formats a non-negative count into 4 FND digits and decimal point flags.
// Counts above 9999 are abbreviated with 'k' or 'm' suffixes.
func formatCount4(n int) ([sideDigitCount]rune, [sideDigitCount]bool) {
	var dots [sideDigitCount]bool

	switch {
	case n < 0:
		return [sideDigitCount]rune{' ', '-', '-', '-'}, dots
	case n <= 9999:
		return asciiToChars4(fmt.Sprintf("%4d", n)), dots
	case n <= 99999:
		// XX.Xk (e.g. 12340 -> "12.3k")
		dots[1] = true
		return asciiToChars4(fmt.Sprintf("%03dk", n/100)), dots
	case n <= 999999:
		// XXXk (e.g. 123400 -> "123k")
		return asciiToChars4(fmt.Sprintf("%3dk", n/1000)), dots
	case n <= 9999999:
		// X.XXm (e.g. 1234000 -> "1.23m")
		dots[0] = true
		return asciiToChars4(fmt.Sprintf("%03dm", n/10000)), dots
	case n <= 99999999:
		// XX.Xm (e.g. 12340000 -> "12.3m")
		dots[1] = true
		return asciiToChars4(fmt.Sprintf("%03dm", n/100000)), dots
	default:
		// XXXm, clamped at 999m
		m := min(n/1000000, 999)
		return asciiToChars4(fmt.Sprintf("%3dm", m)), dots
	}
}

func asciiToChars4(s string) [sideDigitCount]rune {
	var chars [sideDigitCount]rune
	for i := range sideDigitCount {
		if i < len(s) {
			chars[i] = rune(s[i])
		}
	}
	return chars
}
