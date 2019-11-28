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
	"tobw/session"
	//	"golang.org/x/text/encoding/charmap"
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
	log.SetLevel(log.InfoLevel)
	// startup message
	log.Infof("%s (%s) is starting up...\n", APP_NAME, APP_CODE)

	// codepage test stuff that needs to be removed
	// testb := byte(176)
	// resb := charmap.CodePage437.DecodeByte(testb)
	// charmap.CodePage437.NewDecoder(). (need to check this)
	//fmt.Println(string(resb))

	// start telnet listener
	log.Infof("Starting telnet listener on port %d...\n", TELNET_PORT)
    srv, err := net.Listen("tcp", ":5000")
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

	// Close the connection when you're done with it.
	defer func () {
		log.Infof("%s - Disconnected\n", telnetConn.RemoteAddr())
		err := term.Close()
		if err!=nil {
			log.Error(err.Error())
		}
	}()

	// Read a bit of data to let the telnet negotiation finish. Ignore any actual data for now.
	_, _ = telnetConn.Read(buf)


	session.Start(term)
}