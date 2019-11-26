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
	"io"
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

func (t AnsiTerminal) ResizeTerminal(w int, h int) {
	if w > 0 {
		t.columns = w
	}
	if h > 0 {
		t.rows = h
	}
	t.Printf("Terminal dimensions: columns=%d, rows=%d\r\n", t.columns, t.rows)
}

// implement standard formatting functions.
// We don't provide scanf-like input functions. We will develop our own input routines.

func (t AnsiTerminal) Print(a ...interface{}) (n int, err error) {
	s := fmt.Sprint(a...)
	n, err = t.Write([]byte(s))
	return
}

func (t AnsiTerminal) Printf(format string, a ...interface{}) (n int, err error) {
	s := fmt.Sprintf(format, a...)
	n, err = t.Write([]byte(s))
	return
}

func (t AnsiTerminal) Println(a ...interface{}) (n int, err error) {
	s := fmt.Sprintln(a...)
	n, err = t.Write([]byte(s))
	return
}

// color stuff

func (t AnsiTerminal) SetColor(c FGColor, bright bool) {
	if bright{
		t.Printf("\x1B[0;1;%dm", c)
	} else {
		t.Printf("\x1B[0;2;%dm", c)
	}
}

func (t AnsiTerminal) SetFullColor(c FGColor, b BGColor, bright bool) {
	if bright {
		t.Printf("\x1B[0;1;%d;%dm", c,b)
	} else {
		t.Printf("\x1B[0;2;%d;%dm", c,b)

	}
}

func (t AnsiTerminal) ClearEOL() {
	t.Printf("\x1B[K")
}

func (t AnsiTerminal) ClearScreen() {
	t.Printf("\x1B[2J")
}

func (t AnsiTerminal) GotoXY(row int, column int) {
	// index 1-based
	t.Printf("\x1B[%d;%dH", row, column)


}

func (t AnsiTerminal) SetBlink(v bool) {
	if v {
		t.Printf("\x1B[5m")
	} else {
		t.Printf("\x1B[25m")
	}
}

