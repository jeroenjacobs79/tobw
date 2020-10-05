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

package session

import (
	"strings"
	"time"

	"github.com/jeroenjacobs79/tobw/internal/ansiterm"
	"github.com/jeroenjacobs79/tobw/internal/config"
	"github.com/jeroenjacobs79/tobw/internal/monitoring"
	"github.com/jeroenjacobs79/tobw/internal/user"
	"github.com/mdp/qrterminal"
)

type TerminalSession struct {
	//
	Terminal       *ansiterm.AnsiTerminal
	ConnectionType config.ConnectionType
	OriginAddress  string
}

// placeholder for hangup channel, so we can use it anywhere in our package
var hangupChannel chan<- *TerminalSession

func CreateSession(term *ansiterm.AnsiTerminal, conntype config.ConnectionType, origin string) *TerminalSession {
	session := TerminalSession{
		Terminal:       term,
		ConnectionType: conntype,
		OriginAddress:  origin,
	}
	return &session
}

func Start(session *TerminalSession, hangup chan<- *TerminalSession) {
	// this delay seems to help with older DOS-based terminals running in DosBox.
	time.Sleep(1 * time.Second)

	term := session.Terminal
	// make our hangup handler global
	hangupChannel = hangup

	// make sure hangup occurs at the end
	defer func() {
		hangupChannel <- session
	}()

	// set metrics
	monitoring.CurrentConnections.Inc()
	switch session.ConnectionType {
	case config.TCPRaw:
		monitoring.CurrentRawConnections.Inc()
	case config.TCPTelnet:
		monitoring.CurrentTelnetConnections.Inc()
	case config.TCPSSH:
		monitoring.CurrentSSHConnections.Inc()
	}

	// start here
	term.ClearScreen()
	term.GotoXY(1, 1)
	term.SetColor(ansiterm.White, true)
	term.Printf("\nTale of the Black Wyvern - ")
	term.SetColor(ansiterm.Green, false)
	term.Printf("City Square\n")
	term.SetColor(ansiterm.Green, true)
	cols, _ := term.GetTerminalSize()
	var line string
	if cols%2 == 0 {
		line = strings.Repeat("-=", cols/2)
	} else {
		line = strings.Repeat("-=", (cols-1)/2)
	}

	term.Printf("%s\n", line)

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

	adminUser := user.User{}
	adminUser.Username = "test user"
	adminUser.SetPassword("mytestpw")

	term.SetColor(ansiterm.White, false)
	term.Print("\nPlease enter your username: ")
	result, err := term.Input(25, ansiterm.InputUpfirst)
	if err != nil {
		return
	}

	term.Print("\nPlease enter your password: ")
	pwResult, err := term.Input(32, ansiterm.InputPassword)
	if err != nil {
		return
	}

	if adminUser.ValidatePassword(pwResult) {
		term.SendTextFile("ansi/citysquare.ans")
		term.Printf("Welcome %s", result)

	} else {
		term.Println("Password incorrect. Disconnecting...")
	}

	return
}
