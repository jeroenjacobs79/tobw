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
	"github.com/mdp/qrterminal"
	"strings"
	"time"
	"tobw/ansiterm"
)

func Start(term *ansiterm.AnsiTerminal) {
	// this delay seems to help with older DOS-based terminals running in DosBox.
	time.Sleep(1*time.Second)

	// start here
	term.ClearScreen()
	term.GotoXY(1, 1)
	term.SetColor(ansiterm.FG_WHITE, true)
	term.Printf("\nTale of the Black Wyvern - ")
	term.SetColor(ansiterm.FG_GREEN, false)
	term.Printf("City Square\n")
	term.SetColor(ansiterm.FG_GREEN, true)
	cols, _ := term.GetTerminalSize()
	var line string
	if cols%2 == 0 {
		line = strings.Repeat("-=", cols/2)
	} else {
		line = strings.Repeat("-=", (cols-1)/2)
	}

	term.Printf("%s\n", line)
	term.SetColor(ansiterm.FG_GREEN, false)
	term.Printf("People are flooding the square. You try to move to your destination, not afraid to use your elbows in the process...\n\n")
	term.SetColor(ansiterm.FG_WHITE, true)
	term.Printf("Choice your destination:\n\n")

	// menu item
	term.DisplayMenuItem('F', "Explore the forest")
	term.Print("\t\t")
	term.DisplayMenuItem('W', "Weapons 'R' Us")
	term.Println("")

	term.DisplayMenuItem('I', "Cockroach Inn")
	term.Print("\t\t")
	term.DisplayMenuItem('A', "Fashion armour store")
	term.Println()

	term.DisplayMenuItem('B', "Cheat'm and Crook Bank")
	term.Print("\t")
	term.DisplayMenuItem('T', "Training ground")
	term.Println()

	term.DisplayMenuItem('H', "Shaman's Healer Hut")
	term.Print("\t\t")
	term.DisplayMenuItem('Y', "Your stats")
	term.Println()

	term.DisplayMenuItem('K', "Fluffy Kitty Brothel")
	term.Print("\t")
	term.DisplayMenuItem('O', "Who's online?")
	term.Println()

	term.DisplayMenuItem('P', "Post Office")
	term.Print("\t\t\t")
	term.DisplayMenuItem('!', "Log off")
	term.Println()

	term.DisplayMenuItem('C', "City Hall")
	term.Println()

	// qr test
	var qrBuffer strings.Builder
	qrConfig := qrterminal.Config{
		Level:      qrterminal.M,
		Writer:     &qrBuffer,
		HalfBlocks: false,
		BlackChar:  qrterminal.WHITE,
		WhiteChar:  qrterminal.BLACK,
		QuietZone:  1,
	}
	// Taken from google example: https://github.com/google/google-authenticator/wiki/Key-Uri-Format
	qrterminal.GenerateWithConfig("otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example", qrConfig)
	// term.Print(qrBuffer.String())

	term.SetColor(ansiterm.FG_WHITE, false)
	term.Print("\nPlease enter your real name: ")
	result, err := term.Input(25, ansiterm.INPUT_UPFIRST)
	if err != nil {
		return
	}

	term.Println(result)
	return
}
