/*
 * Copyright (c) 2019 Jeroen Jacobs.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 2 as published by
 * the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package ansiterm

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"unicode"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
	"golang.org/x/text/encoding/charmap"
)

type AnsiTerminal struct {
	ioDevice io.ReadWriteCloser
	*bufio.ReadWriter
	columns     int
	rows        int
	Cp437toUtf8 bool
}

type AnsiColor int
type InputMode int

const (
	Reset   AnsiColor = -1
	Black   AnsiColor = 0
	Red     AnsiColor = 1
	Green   AnsiColor = 2
	Yellow  AnsiColor = 3
	Blue    AnsiColor = 4
	Magenta AnsiColor = 5
	Cyan    AnsiColor = 6
	White   AnsiColor = 7

	InputAll      InputMode = 0
	InputDigit    InputMode = 1
	InputUpall    InputMode = 2
	InputPassword InputMode = 3
	InputUpfirst  InputMode = 4
)

func CreateAnsiTerminal(device io.ReadWriteCloser) *AnsiTerminal {
	term := AnsiTerminal{
		ioDevice:   device,
		ReadWriter: bufio.NewReadWriter(bufio.NewReader(device), bufio.NewWriter(device)),
		// this is pretty standard in case we don't receive any updates on the size
		columns: 80,
		rows:    24,
	}
	return &term
}

func (t *AnsiTerminal) Close() (err error) {
	t.Flush()
	return t.ioDevice.Close()
}

func (t *AnsiTerminal) WriteText(data []byte) (totalWritten int, err error) {
	if t.Cp437toUtf8 {
		result, err := charmap.CodePage437.NewDecoder().Bytes(data)
		if err == nil {
			totalWritten, err = t.Write(result)
		}
	} else {
		totalWritten, err = t.Write(data)
	}
	return
}

func (t *AnsiTerminal) ResizeTerminal(w int, h int) {
	if w > 0 {
		t.columns = w
	}
	if h > 0 {
		t.rows = h
	}
}

func (t *AnsiTerminal) GetTerminalSize() (columns int, rows int) {
	return t.columns, t.rows
}

// implement standard formatting functions.
// We don't provide scanf-like input functions. We will develop our own input routines.

func (t *AnsiTerminal) Print(a ...interface{}) (n int, err error) {
	s := fmt.Sprint(a...)
	final := strings.NewReplacer("\r\n", "\r\n", "\n", "\r\n").Replace(s)
	n, err = t.WriteText([]byte(final))
	_ = t.Flush()
	return
}

func (t *AnsiTerminal) Printf(format string, a ...interface{}) (n int, err error) {
	s := fmt.Sprintf(format, a...)
	final := strings.NewReplacer("\r\n", "\r\n", "\n", "\r\n").Replace(s)
	n, err = t.WriteText([]byte(final))
	_ = t.Flush()
	return
}

func (t *AnsiTerminal) Println(a ...interface{}) (n int, err error) {
	s := fmt.Sprintln(a...)
	final := strings.NewReplacer("\r\n", "\r\n", "\n", "\r\n").Replace(s)
	n, err = t.WriteText([]byte(final))
	_ = t.Flush()
	return
}

// color stuff

func (t *AnsiTerminal) SetColor(fg AnsiColor, bright bool) {
	fgColor := convertColorToFGValue(fg)
	if bright {
		t.Printf("\x1B[0;1;%dm", fgColor)
	} else {
		t.Printf("\x1B[0;22;%dm", fgColor)
	}
}

func (t *AnsiTerminal) SetFullColor(fg AnsiColor, bg AnsiColor, bright bool) {
	fgColor := convertColorToFGValue(fg)
	bgColor := convertColorToBGValue(bg)
	if bright {
		t.Printf("\x1B[0;1;%d;%dm", fgColor, bgColor)
	} else {
		t.Printf("\x1B[0;2;%d;%dm", fgColor, bgColor)

	}
}

func convertColorToFGValue(color AnsiColor) (result byte) {
	if color == Reset {
		result = 0
	} else {
		result = byte(color) + 30
	}
	return
}

func convertColorToBGValue(color AnsiColor) (result byte) {
	if color == Reset {
		result = 0
	} else {
		result = byte(color) + 40
	}
	return
}

func (t *AnsiTerminal) ClearEOL() {
	t.Printf("\x1B[K")
}

func (t *AnsiTerminal) ClearScreen() {
	t.Printf("\x1B[2J")
}

func (t *AnsiTerminal) GotoXY(row int, column int) {
	// index 1-based
	t.Printf("\x1B[%d;%dH", row, column)

}

func (t *AnsiTerminal) SetBlink(v bool) {
	if v {
		t.Printf("\x1B[5m")
	} else {
		t.Printf("\x1B[25m")
	}
}

func (t *AnsiTerminal) DisplayMenuItem(id rune, description string) {
	t.SetColor(Green, false)
	t.Printf("[")
	t.SetColor(Red, true)
	t.Printf("%c", id)
	t.SetColor(Green, false)
	t.Printf("] %s", description)

}

// Input routines

func (t *AnsiTerminal) WaitKey(ignoreCase bool) (r rune, err error) {
	// wait for key and return. If key is character, it is converted to uppercase.
	found := false
	countRead := 0
	for !found {
		buffer := make([]byte, 256)
		countRead, err = t.Read(buffer)
		//		log.Traceln(buffer[:countRead])
		for _, value := range string(buffer[:countRead]) {
			if unicode.IsLower(value) && ignoreCase {
				r = unicode.ToUpper(value)
			} else {
				r = value
			}
			found = true
			break
		}
		if err != nil {
			break
		}
	}
	return
}

func (t *AnsiTerminal) WaitKeys(allowed string, ignoreCase bool) (r rune, err error) {
	// wait for key that is permitted and return. If key is character, it is converted to uppercase.
	found := false
	countRead := 0
	for !found {
		buffer := make([]byte, 256)
		countRead, err = t.Read(buffer)
		log.Traceln(buffer[:countRead])
		for _, value := range string(buffer[:countRead]) {
			var current rune
			if unicode.IsLower(value) && ignoreCase {
				current = unicode.ToUpper(value)
			} else {
				current = value
			}
			for _, allowedRune := range allowed {
				if unicode.IsLower(allowedRune) && ignoreCase {
					allowedRune = unicode.ToUpper(allowedRune)
				}
				if current == allowedRune {
					found = true
					r = current
					break
				}
			}
		}
		if err != nil {
			break
		}
	}
	return
}

func (t *AnsiTerminal) Input(size int, mode InputMode) (result string, err error) {
	// print field
	var inputBuffer strings.Builder
	var inputCounter = 0
	var ch rune
	var lastChar rune
	t.SetFullColor(Blue, Blue, false)
	for i := 0; i < size; i++ {
		t.Print(" ")
	}
	t.Printf("\x1B[%dD", size)
	t.SetFullColor(White, Blue, false)

	// We have drawn our input box, now let's get input
	ch, err = t.WaitKey(false)
	for ch != '\r' {
		if err != nil {
			break
		}
		switch ch {
		case '\b', '\u007F':
			// backspace entered
			if inputCounter > 0 {
				// the fact string-slicing is byte-based and not rune-based, is a PITA.
				temp := inputBuffer.String()
				_, sizeLast := utf8.DecodeLastRuneInString(temp)
				newValue := temp[:len(temp)-sizeLast]
				inputBuffer.Reset()
				inputBuffer.WriteString(newValue)
				inputCounter--
				t.Printf("%c %c", '\b', '\b')

			}

		default:
			if inputCounter < size {
				if unicode.IsPrint(ch) {
					switch mode {
					case InputAll:
						inputBuffer.WriteRune(ch)
						inputCounter++
						t.Printf("%c", ch)
					case InputPassword:
						inputBuffer.WriteRune(ch)
						inputCounter++
						t.Print("*")
					case InputUpall:
						if unicode.IsLower(ch) {
							ch = unicode.ToUpper(ch)
						}
						inputBuffer.WriteRune(ch)
						inputCounter++
						t.Printf("%c", ch)
					case InputDigit:
						if unicode.IsDigit(ch) {
							inputBuffer.WriteRune(ch)
							inputCounter++
							t.Printf("%c", ch)
						}
					case InputUpfirst:
						if unicode.IsSpace(lastChar) || inputCounter == 0 {
							ch = unicode.ToUpper(ch)
						}
						lastChar = ch
						inputBuffer.WriteRune(ch)
						inputCounter++
						t.Printf("%c", ch)
					}

				}
			}
		}
		// next char
		ch, err = t.WaitKey(false)
	}
	t.Printf("\x1B[%dC", size-inputCounter)
	t.Print("\n")
	t.Printf("\x1B[0m")
	result = inputBuffer.String()
	return
}

func (t *AnsiTerminal) SendTextFile(path string) {
	privateBytes, err := ioutil.ReadFile(path)
	if err == nil {
		t.WriteText(privateBytes)
		_ = t.Flush()
	}
}
