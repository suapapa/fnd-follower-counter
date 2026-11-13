package display

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/suapapa/go_devices/tm1638"
)

const totalDigits = sideDigitCount * 2

// sevenSegFont mirrors tm1638's 7-segment bit mapping:
//
//	 -- 0 --
//	|       |
//	5       1
//	 -- 6 --
//	4       2
//	|       |
//	 -- 3 --  .7
var sevenSegFont = map[rune]byte{
	' ': 0b00000000,
	'-': 0b01000000,
	'.': 0b10000000,
	'0': 0b00111111,
	'1': 0b00000110,
	'2': 0b01011011,
	'3': 0b01001111,
	'4': 0b01100110,
	'5': 0b01101101,
	'6': 0b01111101,
	'7': 0b00100111,
	'8': 0b01111111,
	'9': 0b01101111,
	'A': 0b01110111,
	'B': 0b01111111,
	'C': 0b00111001,
	'D': 0b00111111,
	'E': 0b01111001,
	'F': 0b01110001,
	'G': 0b00111101,
	'H': 0b01110110,
	'I': 0b00000110,
	'J': 0b00011111,
	'K': 0b01101001,
	'L': 0b00111000,
	'M': 0b00010101,
	'N': 0b00110111,
	'O': 0b00111111,
	'P': 0b01110011,
	'Q': 0b01100111,
	'R': 0b00110001,
	'S': 0b01101101,
	'T': 0b01111000,
	'U': 0b00111110,
	'V': 0b00101010,
	'W': 0b00011101,
	'X': 0b01110110,
	'Y': 0b01101110,
	'Z': 0b01011011,
	'a': 0b01011111,
	'b': 0b01111100,
	'c': 0b01011000,
	'd': 0b01011110,
	'e': 0b01111011,
	'f': 0b00110001,
	'g': 0b01101111,
	'h': 0b01110100,
	'i': 0b00000100,
	'j': 0b00001110,
	'k': 0b01110101,
	'l': 0b00110000,
	'm': 0b01010101,
	'n': 0b01010100,
	'o': 0b01011100,
	'p': 0b01110011,
	'q': 0b01100111,
	'r': 0b01010000,
	's': 0b01101101,
	't': 0b01111000,
	'u': 0b00011100,
	'v': 0b00101010,
	'w': 0b00011101,
	'x': 0b01110110,
	'y': 0b01101110,
	'z': 0b01000111,
	'_': 0b00001000,
}

// ConsoleFND buffers 8 FND digits and LEDs and renders them as 7-segment ASCII art.
type ConsoleFND struct {
	out   io.Writer
	chars [totalDigits]rune
	dots  [totalDigits]bool
	leds  [totalDigits]tm1638.Color
}

// NewConsoleFND creates a ConsoleFND that writes ASCII-art frames to out.
func NewConsoleFND(out io.Writer) *ConsoleFND {
	c := &ConsoleFND{out: out}
	for i := range c.chars {
		c.chars[i] = ' '
	}
	return c
}

// SetChar sets the character and decimal point for digit pos.
func (c *ConsoleFND) SetChar(pos int, ch rune, dot bool) {
	if pos < 0 || pos >= totalDigits {
		return
	}
	c.chars[pos] = ch
	c.dots[pos] = dot
}

// SetLed sets the LED color above digit pos.
func (c *ConsoleFND) SetLed(pos int, led tm1638.Color) {
	if pos < 0 || pos >= totalDigits {
		return
	}
	c.leds[pos] = led
}

// SetString writes str across the FND digits starting at position 0.
func (c *ConsoleFND) SetString(str string) {
	i := 0
	for _, r := range str {
		if i >= totalDigits {
			break
		}
		c.SetChar(i, r, false)
		i++
	}
}

// Flush writes the current 7-segment ASCII-art frame to the configured writer.
func (c *ConsoleFND) Flush() {
	if c.out == nil {
		return
	}
	if _, err := fmt.Fprint(c.out, c.Format()); err != nil {
		log.Printf("write console FND: %v", err)
	}
}

// Format renders the buffered LEDs and 8 7-segment digits as an ASCII-art box.
func (c *ConsoleFND) Format() string {
	var ledRow, topRow, midRow, botRow strings.Builder

	for i := range totalDigits {
		if i > 0 {
			gap := " "
			if i == sideDigitCount {
				gap = "  "
			}
			ledRow.WriteString(gap)
			topRow.WriteString(gap)
			midRow.WriteString(gap)
			botRow.WriteString(gap)
		}

		ledRow.WriteString(formatLEDCell(c.leds[i]))
		top, mid, bot := formatSevenSegDigit(c.chars[i], c.dots[i])
		topRow.WriteString(top)
		midRow.WriteString(mid)
		botRow.WriteString(bot)
	}

	border := "+" + strings.Repeat("-", 42) + "+"
	var out strings.Builder
	out.WriteString(border + "\n")
	out.WriteString("| " + ledRow.String() + " |\n")
	out.WriteString("| " + topRow.String() + " |\n")
	out.WriteString("| " + midRow.String() + " |\n")
	out.WriteString("| " + botRow.String() + " |\n")
	out.WriteString(border + "\n")
	return out.String()
}

func formatLEDCell(led tm1638.Color) string {
	switch led {
	case tm1638.Green:
		return "(G) "
	case tm1638.Red:
		return "(R) "
	default:
		return "( ) "
	}
}

func formatSevenSegDigit(ch rune, dot bool) (string, string, string) {
	mask := sevenSegFont[ch]
	if mask&0x80 != 0 {
		dot = true
	}

	seg := func(bit uint, glyph string) string {
		if mask&(1<<bit) != 0 {
			return glyph
		}
		return " "
	}

	dotGlyph := " "
	if dot {
		dotGlyph = "."
	}

	top := " " + seg(0, "_") + "  "
	mid := seg(5, "|") + seg(6, "_") + seg(1, "|") + " "
	bot := seg(4, "|") + seg(3, "_") + seg(2, "|") + dotGlyph
	return top, mid, bot
}
