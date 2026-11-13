package display

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/suapapa/go_devices/tm1638"
)

type mockFND struct {
	chars   [8]rune
	dots    [8]bool
	leds    [8]tm1638.Color
	strings []string
}

func (m *mockFND) SetChar(pos int, c rune, dot bool) {
	m.chars[pos] = c
	m.dots[pos] = dot
}

func (m *mockFND) SetLed(pos int, led tm1638.Color) {
	m.leds[pos] = led
}

func (m *mockFND) SetString(str string) {
	m.strings = append(m.strings, str)
}

func TestFormatCount4(t *testing.T) {
	tests := []struct {
		n        int
		wantChar [4]rune
		wantDot  [4]bool
	}{
		{
			n:        0,
			wantChar: [4]rune{' ', ' ', ' ', '0'},
			wantDot:  [4]bool{false, false, false, false},
		},
		{
			n:        42,
			wantChar: [4]rune{' ', ' ', '4', '2'},
			wantDot:  [4]bool{false, false, false, false},
		},
		{
			n:        9999,
			wantChar: [4]rune{'9', '9', '9', '9'},
			wantDot:  [4]bool{false, false, false, false},
		},
		{
			n:        10000,
			wantChar: [4]rune{'1', '0', '0', 'k'},
			wantDot:  [4]bool{false, true, false, false},
		},
		{
			n:        12340,
			wantChar: [4]rune{'1', '2', '3', 'k'},
			wantDot:  [4]bool{false, true, false, false},
		},
		{
			n:        123400,
			wantChar: [4]rune{'1', '2', '3', 'k'},
			wantDot:  [4]bool{false, false, false, false},
		},
		{
			n:        1234000,
			wantChar: [4]rune{'1', '2', '3', 'm'},
			wantDot:  [4]bool{true, false, false, false},
		},
		{
			n:        12340000,
			wantChar: [4]rune{'1', '2', '3', 'm'},
			wantDot:  [4]bool{false, true, false, false},
		},
		{
			n:        123400000,
			wantChar: [4]rune{'1', '2', '3', 'm'},
			wantDot:  [4]bool{false, false, false, false},
		},
		{
			n:        2000000000,
			wantChar: [4]rune{'9', '9', '9', 'm'},
			wantDot:  [4]bool{false, false, false, false},
		},
	}

	for _, tc := range tests {
		gotChar, gotDot := formatCount4(tc.n)
		if gotChar != tc.wantChar || gotDot != tc.wantDot {
			t.Errorf(
				"formatCount4(%d) = (%q, %v), want (%q, %v)",
				tc.n,
				string(gotChar[:]),
				gotDot,
				string(tc.wantChar[:]),
				tc.wantDot,
			)
		}
	}
}

func TestRender(t *testing.T) {
	dev := &mockFND{}
	left := SideState{Count: 12340, HasValue: true}
	right := SideState{Count: 0, HasValue: false, Err: errors.New("boom")}

	Render(dev, left, right)

	wantChars := [8]rune{'1', '2', '3', 'k', ' ', 'E', 'r', 'r'}
	if dev.chars != wantChars {
		t.Fatalf("chars = %q, want %q", string(dev.chars[:]), string(wantChars[:]))
	}
	wantDots := [8]bool{false, true, false, true, false, false, false, false}
	if dev.dots != wantDots {
		t.Fatalf("dots = %v, want %v", dev.dots, wantDots)
	}
	wantLEDs := [8]tm1638.Color{
		tm1638.Green, tm1638.Green, tm1638.Green, tm1638.Green,
		tm1638.Red, tm1638.Red, tm1638.Red, tm1638.Red,
	}
	if dev.leds != wantLEDs {
		t.Fatalf("leds = %v, want %v", dev.leds, wantLEDs)
	}
}

func TestConsoleFND(t *testing.T) {
	var buf bytes.Buffer
	dev := NewConsoleFND(&buf)

	left := SideState{Count: 12340, HasValue: true}
	right := SideState{Count: 0, HasValue: false, Err: errors.New("boom")}

	Render(dev, left, right)

	want := strings.Join([]string{
		"+------------------------------------------+",
		"| (G)  (G)  (G)  (G)   (R)  (R)  (R)  (R)  |",
		"|       _    _    _          _             |",
		"|   |   _|   _|  |_         |_    _    _   |",
		"|   |  |_ .  _|  | |.       |_   |    |    |",
		"+------------------------------------------+",
		"",
	}, "\n")

	if got := buf.String(); got != want {
		t.Fatalf("console output mismatch:\ngot:\n%s\nwant:\n%s", got, want)
	}
}
