/*
 * Copyright (c) 2019 Head In Cloud BVBA.
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
 *
 *
 */

package ansiterm

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"strings"
	"unicode"
)

type AnsiTerminal struct {
	io.ReadWriteCloser
	columns int
	rows int
}

type FGColor byte
type BGColor byte

const (

	FG_BLACK FGColor 	= 30
	FG_RED FGColor 		= 31
	FG_GREEN FGColor 	= 32
	FG_YELLOW FGColor	= 33
	FG_BLUE FGColor 	= 34
	FG_MAGENTA FGColor 	= 35
	FG_CYAN FGColor 	= 36
	FG_WHITE FGColor 	= 37

	BG_BLACK BGColor 	= 40
	BG_RED BGColor 		= 41
	BG_GREEN BGColor 	= 42
	BG_YELLOW BGColor	= 43
	BG_BLUE BGColor 	= 44
	BG_MAGENTA BGColor 	= 45
	BG_CYAN BGColor 	= 46
	BG_WHITE BGColor 	= 47


)


func CreateAnsiTerminal(device io.ReadWriteCloser) (*AnsiTerminal){
	term := AnsiTerminal{
		ReadWriteCloser: device,
		// this is pretty standard in case we don't receive any updates on the size
		columns: 80,
		rows: 24,
	}
	return &term
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
	final:= strings.NewReplacer("\r\n", "\r\n", "\n", "\r\n").Replace(s)
	n, err = t.Write([]byte(final))
	return
}

func (t *AnsiTerminal) Printf(format string, a ...interface{}) (n int, err error) {
	s := fmt.Sprintf(format, a...)
	final:= strings.NewReplacer("\r\n", "\r\n", "\n", "\r\n").Replace(s)
	n, err = t.Write([]byte(final))
	return
}

func (t *AnsiTerminal) Println(a ...interface{}) (n int, err error) {
	s := fmt.Sprintln(a...)
	final:= strings.NewReplacer("\r\n", "\r\n", "\n", "\r\n").Replace(s)
	n, err = t.Write([]byte(final))
	return
}

// color stuff

func (t *AnsiTerminal) SetColor(c FGColor, bright bool) {
	if bright{
		t.Printf("\x1B[0;1;%dm", c)
	} else {
		t.Printf("\x1B[0;22;%dm", c)
	}
}

func (t *AnsiTerminal) SetFullColor(c FGColor, b BGColor, bright bool) {
	if bright {
		t.Printf("\x1B[0;1;%d;%dm", c,b)
	} else {
		t.Printf("\x1B[0;2;%d;%dm", c,b)

	}
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

// Input routines

func (t *AnsiTerminal) WaitKey() (r rune, err error) {
	// wait for key and return. If key is character, it is converted to uppercase.
	found := false
	countRead := 0
	for !found {
		buffer := make([]byte, 256)
		countRead, err = t.Read(buffer)
		log.Traceln(buffer[:countRead])
		for _, value := range string(buffer[:countRead]) {
			if unicode.IsLower(value) {
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

func (t *AnsiTerminal) WaitKeys(allowed string) (r rune, err error) {
	// wait for key that is permitted and return. If key is character, it is converted to uppercase.
	found := false
	countRead := 0
	for !found {
		buffer := make([]byte, 256)
		countRead, err = t.Read(buffer)
		log.Traceln(buffer[:countRead])
		for _, value := range string(buffer[:countRead]) {
			var current rune
			if unicode.IsLower(value) {
				current = unicode.ToUpper(value)
			} else {
				current = value
			}
			for _, allowedRune := range allowed {
				if unicode.IsLower(allowedRune) {
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

func (t *AnsiTerminal) Input(size int) (result string, err error) {
	// print field
	var inputBuffer strings.Builder
	var inputCounter = 0
	var ch rune
	t.SetFullColor(FG_BLUE, BG_BLUE, false)
	for i := 0; i < size; i++ {
		t.Print(" ")
	}
	t.Printf("\x1B[%dD", size)
	t.SetFullColor(FG_WHITE, BG_BLUE, false)

	// We have drawn our input box, now let's get input
	ch, err = t.WaitKey()
	for ch != '\r' {
		if err != nil {
			break
		}
		switch ch {
		case '\a', '\f', '\n', '\t', '\v', '\\', '\'', '"':
			// ignore these
		case '\b', '\u007F':
			// backspace entered
			t.Printf("%c", ch)

		default:
			if inputCounter < size {
				t.Printf("%c", ch)
				inputCounter++
			}
		}
		// next char
		ch, err = t.WaitKey()
	}
	result = inputBuffer.String()
	return
}