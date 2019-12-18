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

package termserve

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"sync"
	"tobw/internal/ansiterm"
	"tobw/internal/session"
	"tobw/internal/telnet"
)

type ConnectionType int

const (
	Telnet ConnectionType = iota
	RawTCP
	Ssh
)

func (t ConnectionType) String() (result string) {
	switch t {
	case Telnet:
		result = "telnet"

	case RawTCP:
		result = "raw"

	case Ssh:
		result = "ssh"
	default:
		result = "unknown"
	}
	return
}

func StartListener(wg *sync.WaitGroup, address string, c ConnectionType, cp437ToUtf8 bool) {
	// start telnet listener
	log.Infof("Starting %s listener on address %s...", c, address)

	// listeners for telnet and ssh
	if c != Ssh {
		srv, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal(err.Error())
		}
		// close listener on exit
		defer func() {
			err := srv.Close()
			if err != nil {
				log.Errorln(err.Error())
			}
			wg.Done()
		}()

		// start accepting connections
		log.Infof("Started %s listener successfully on address %s.", c, address)
		for {
			// Listen for an incoming connection.
			conn, err := srv.Accept()
			if err != nil {
				log.Errorln(err.Error())
			}
			// Handle connections in a new goroutine.
			switch c {
			case Telnet:
				go handleTelnetRequest(conn, cp437ToUtf8)
			case RawTCP:
				go handleRawRequest(conn, cp437ToUtf8)
			}

		}
	} else {
		// Ssh is more complicated. We disable authentication, as this is handled within the game.
		// We also configure the private key here.
		// This adapted from the example here: https://godoc.org/golang.org/x/crypto/ssh#example-NewServerConn
		// This works totally different from the telnet/raw implementation.

		config := &ssh.ServerConfig{
			NoClientAuth: true,
		}
		privateBytes, err := ioutil.ReadFile("/Users/jeroenjacobs/.ssh/tobw_rsa")
		if err != nil {
			log.Fatalln("Failed to load private key:", err)
		}
		privateKey, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			log.Fatalln("Failed to parse private key:", err)
		}
		config.AddHostKey(privateKey)

		// start our actual listener
		srv, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal(err.Error())
		}
		// close listener on exit
		defer func() {
			err := srv.Close()
			if err != nil {
				log.Errorln(err.Error())
			}
			wg.Done()
		}()

		// start accepting connections
		log.Infof("Started %s listener successfully on address %s.", c, address)

		for {
			// Listen for an incoming connection.
			conn, err := srv.Accept()
			if err != nil {
				log.Errorln(err.Error())
			}

			go handleSshRequest(conn, config, cp437ToUtf8)
		}
	}
}

func handleSshRequest(conn net.Conn, conf *ssh.ServerConfig, cp437ToUtf8 bool) {
	log.Infof("%s - Connected", conn.RemoteAddr())
	_, chans, reqs, err := ssh.NewServerConn(conn, conf)
	if err != nil {
		log.Errorf("%s - ssh handshake failed", conn.RemoteAddr())
	} else {
		log.Tracef("%s - ssh handshake successful", conn.RemoteAddr())
	}

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			err = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			if err != nil {
				log.Errorln(err.Error())
			}
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Errorln(err.Error())
		}

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				err = req.Reply(req.Type == "pty-req" || req.Type == "shell", nil)
				if err != nil {
					log.Errorln(err.Error())
				}
			}
		}(requests)

		term := ansiterm.CreateAnsiTerminal(channel)
		term.Cp437toUtf8 = cp437ToUtf8
		session.Start(term)
		log.Infof("%s - Disconnected", conn.RemoteAddr())
		err = term.Close()
		if err != nil {
			log.Error(err.Error())
		}

	}
}

func handleTelnetRequest(conn net.Conn, cp437ToUtf8 bool) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	telnetConn := telnet.NewConnection(conn)
	log.Infof("%s - Connected", telnetConn.RemoteAddr())
	term := ansiterm.CreateAnsiTerminal(telnetConn)
	term.Cp437toUtf8 = cp437ToUtf8
	telnetConn.InstallResizeHandler(term.ResizeTerminal)
	telnetConn.RequestTermSize()
	log.Traceln(term)

	// Close the connection when you're done with it.
	defer func() {
		log.Infof("%s - Disconnected", telnetConn.RemoteAddr())
		err := term.Close()
		if err != nil {
			log.Error(err.Error())
		}
	}()

	// Read a bit of data to let the telnet negotiation finish. Ignore any actual data for now.
	telnetConn.Read(buf)

	session.Start(term)
}

func handleRawRequest(conn net.Conn, cp437ToUtf8 bool) {
	log.Infof("%s - Connected", conn.RemoteAddr())
	term := ansiterm.CreateAnsiTerminal(conn)
	term.Cp437toUtf8 = cp437ToUtf8

	// Close the connection when you're done with it.
	defer func() {
		log.Infof("%s - Disconnected", conn.RemoteAddr())
		err := term.Close()
		if err != nil {
			log.Error(err.Error())
		}
	}()
	session.Start(term)
}