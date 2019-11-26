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

package main

import (
	log "github.com/sirupsen/logrus"
	"tobw/ansiterm"

	//	"golang.org/x/text/encoding/charmap"
	"io"
	"net"
	"tobw/telnet"
)

const (
	APP_NAME = "Tale of the Black Wyvern"
	APP_CODE = "TOBW"
	TELNET_PORT = 5000
)

func main() {
	// set log format to include timestamp, even when TTY is attached.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:true},
	)
	// set log level
	log.SetLevel(log.DebugLevel)
	// startup message
	log.Infof("%s (%s) is starting up...\n", APP_NAME, APP_CODE)

	// codepage test stuff that needs to be removed
	// testb := byte(176)
	// resb := charmap.CodePage437.DecodeByte(testb)
	// charmap.CodePage437.NewDecoder(). (need to check this)
	//fmt.Println(string(resb))

	// start telnet listener
	log.Infof("Starting telnet listener on port %d...\n", TELNET_PORT)
    srv, err := net.Listen("tcp", "127.0.0.1:5000")
	if err != nil {
		log.Fatal(err.Error())
	}
	// close listener on exit
	defer func () {
		err := srv.Close()
		if err!=nil {
			log.Errorln(err.Error())
		}
	}()

    // start accepting connections
	log.Infof("Telnet listener started.")
	for {
		// Listen for an incoming connection.
		conn, err := srv.Accept()
		if err != nil {
			log.Errorln(err.Error())
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	telnetConn := telnet.NewConnection(conn)
	log.Infof("%s - Connected\n", telnetConn.RemoteAddr())
	term := ansiterm.CreateAnsiTerminal(telnetConn)
	telnetConn.InstallResizeHandler(term.ResizeTerminal)
	telnetConn.RequestTermSize()
	log.Traceln(term)
	active := true

	// Close the connection when you're done with it.
	defer func () {
		log.Infof("%s - Disconnected\n", telnetConn.RemoteAddr())
		err := term.Close()
		if err!=nil {
			log.Error(err.Error())
		}
	}()

	// print our welcome header
	term.ClearScreen()
	term.GotoXY(1,1)
	term.SetColor(ansiterm.FG_BLACK, false)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_RED, false)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_GREEN, false)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_YELLOW, false)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_BLUE, false)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_MAGENTA, false)
	term.SetBlink(true)

	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_CYAN, false)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_WHITE, false)
	term.Printf("Welcome to our game!\n\r")

	term.GotoXY(13,1)
	term.SetColor(ansiterm.FG_BLACK, true)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_RED, true)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_GREEN, true)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_YELLOW, true)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_BLUE, true)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_MAGENTA, true)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_CYAN, true)
	term.Printf("Welcome to our game!\n\r")
	term.SetColor(ansiterm.FG_WHITE, true)
	term.Printf("Welcome to our game!\n\r")

	term.GotoXY(1,40)
	term.SetFullColor(ansiterm.FG_BLACK, ansiterm.BG_BLUE, false)
	term.Printf("Welcome to our game!\n\r")

	// while connection is active, process event loop
	for active {
		// Read the incoming connection into the buffer.
		reqLen, err := telnetConn.Read(buf)
		if reqLen > 0 {
			_, err := telnetConn.Write(buf[:reqLen])
			if err != nil {
				log.Errorln(err.Error())
			}
		}
		if err != nil {
			switch err {
			case io.EOF:
				// do nothing, as this is expected when a connection is closed by the client
			default:
				log.Errorf("%s - %s\n", telnetConn.RemoteAddr(), err.Error())
			}
			active = false
		}
	}
}