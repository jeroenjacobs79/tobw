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

package session

import (
	"strings"
	"tobw/ansiterm"
)

func Start(term *ansiterm.AnsiTerminal) {
	term.ClearScreen()
	term.GotoXY(1,1)
	term.SetColor(ansiterm.FG_WHITE, true)
	term.Printf("\nTale of the Black Wyvern - ")
	term.SetColor(ansiterm.FG_GREEN, false)
	term.Printf("City Square\n")
	term.SetColor(ansiterm.FG_GREEN, true)
	cols, _ := term.GetTerminalSize()
	var line string
	if (cols%2 == 0) {
		line = strings.Repeat("-=",cols/2)
	} else {
		line = strings.Repeat("-=",(cols-1)/2)
	}

	term.Printf("%s\n", line)
	term.SetColor(ansiterm.FG_GREEN, false)
	term.Printf("People are flooding the square. You try to move to your destination, not afraid to use your elbows in the process...\n\n")
	term.SetColor(ansiterm.FG_WHITE, true)
	term.Printf("Choice your destination:\n\n")

	// menu item
	term.DisplayMenuItem('F', "Explore the forest")
	term.Println()

	term.DisplayMenuItem('I', "Cockroach Inn")
	term.Println()

	term.DisplayMenuItem('B', "Cheat'm and Crook Bank")
	term.Println()

	term.DisplayMenuItem('H', "Shaman's Healer Hut")
	term.Println()

	term.DisplayMenuItem('F', "Fluffy Kitty Brothel")
	term.Println()

	term.SetColor(ansiterm.FG_WHITE, false)
	term.Print("\nPlease enter your real name: ")
	result, err := term.Input(25, ansiterm.INPUT_UPFIRST)
	if err != nil {
		return
	}

	term.Println(result)
	return
}
