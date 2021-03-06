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

package termserve

import (
	"encoding/binary"
	"io/ioutil"
	"net"
	"sync"

	"github.com/jeroenjacobs79/tobw/internal/ansiterm"
	"github.com/jeroenjacobs79/tobw/internal/config"
	"github.com/jeroenjacobs79/tobw/internal/monitoring"
	"github.com/jeroenjacobs79/tobw/internal/session"
	"github.com/jeroenjacobs79/tobw/internal/telnet"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var (
	hangupChannel chan *session.TerminalSession
)

func init() {
	// create our channels
	hangupChannel = make(chan *session.TerminalSession)
	go hangupTerminalSession(hangupChannel)
}

func hangupTerminalSession(sessions <-chan *session.TerminalSession) {
	for cleanedSession := range sessions {
		err := cleanedSession.Terminal.Close()
		log.Infof("%s - Disconnected", cleanedSession.OriginAddress)
		monitoring.CurrentConnections.Dec()
		switch cleanedSession.ConnectionType {
		case config.TCPRaw:
			monitoring.CurrentRawConnections.Dec()
		case config.TCPTelnet:
			monitoring.CurrentTelnetConnections.Dec()
		case config.TCPSSH:
			monitoring.CurrentSSHConnections.Dec()
		default:
			log.Errorf("Unknown terminal-type detected!")
		}
		if err != nil {
			log.Error(err.Error())
		}
		// clear pointer so objects can be garbage-collected
		cleanedSession.Terminal = nil
		cleanedSession = nil
	}
}

func StartListener(wg *sync.WaitGroup, address string, c config.ConnectionType, cp437ToUtf8 bool) {
	// start telnet listener
	log.Infof("Starting %s listener on address %s...", c, address)

	// listeners for telnet and raw tcp
	if c != config.TCPSSH {
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
			case config.TCPTelnet:
				go handleTelnetRequest(conn, cp437ToUtf8)
			case config.TCPRaw:
				go handleRawRequest(conn, cp437ToUtf8)
			}

		}
	} else {
		// Ssh is more complicated. We disable authentication, as this is handled within the game.
		// We also configure the private key here.
		// This adapted from the example here: https://godoc.org/golang.org/x/crypto/ssh#example-NewServerConn
		// This works totally different from the telnet/raw implementation.

		sshConfig := &ssh.ServerConfig{
			NoClientAuth: true,
		}
		privateBytes, err := ioutil.ReadFile(config.AppOptions.SSHPrivateKey)
		if err != nil {
			log.Fatalln("Failed to load private key:", err)
		}
		privateKey, err := ssh.ParsePrivateKey(privateBytes)
		if err != nil {
			log.Fatalln("Failed to parse private key:", err)
		}
		sshConfig.AddHostKey(privateKey)

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

			go handleSSHRequest(conn, sshConfig, cp437ToUtf8)
		}
	}
}

// ssh connection handling

func parseSize(data []byte) (w uint32, h uint32) {
	w = binary.BigEndian.Uint32(data)
	h = binary.BigEndian.Uint32(data[4:])
	return
}

func handleSSHRequest(conn net.Conn, conf *ssh.ServerConfig, cp437ToUtf8 bool) {
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

		term := ansiterm.CreateAnsiTerminal(channel)
		term.Cp437toUtf8 = cp437ToUtf8
		currentSession := session.CreateSession(term, config.TCPSSH, conn.RemoteAddr().String())

		// Sessions have out-of-band requests such as "shell",
		// "pty-req" and "env".  Here we handle only the
		// "shell" request.
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "pty-req":
					termLength := int(req.Payload[3]) // this is very naive, as the actual length is encoded as vlint32: http://lists.w3.org/Archives/Public/ietf-tls/msg02555.html
					w, h := parseSize(req.Payload[termLength+4 : termLength+12])
					log.Debugf("%s - receive pty-request for terminal size w:%d h:%d", conn.RemoteAddr(), w, h)
					term.ResizeTerminal(int(w), int(h))
					ok = true
				case "window-change":
					w, h := parseSize(req.Payload[:8])
					log.Debugf("%s - receive window-change for terminal size w:%d h:%d", conn.RemoteAddr(), w, h)
					term.ResizeTerminal(int(w), int(h))
				case "shell":
					ok = true

				}
				if req.WantReply {
					err = req.Reply(ok, nil)
					if err != nil {
						log.Errorln(err.Error())
					}
				}
			}
		}(requests)

		// start actual session
		session.Start(currentSession, hangupChannel)

	}
}

// telnet connection handling

func handleTelnetRequest(conn net.Conn, cp437ToUtf8 bool) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	telnetConn := telnet.NewConnection(conn)
	log.Infof("%s - Connected", telnetConn.RemoteAddr())
	term := ansiterm.CreateAnsiTerminal(telnetConn)
	term.Cp437toUtf8 = cp437ToUtf8
	currentSession := session.CreateSession(term, config.TCPTelnet, conn.RemoteAddr().String())
	telnetConn.InstallResizeHandler(term.ResizeTerminal)
	telnetConn.RequestTermSize()
	log.Traceln(term)

	// Read a bit of data to let the telnet negotiation finish. Ignore any actual data for now.
	_, _ = telnetConn.Read(buf)

	session.Start(currentSession, hangupChannel)
}

// Raw TCP connection handling

func handleRawRequest(conn net.Conn, cp437ToUtf8 bool) {
	log.Infof("%s - Connected", conn.RemoteAddr())
	term := ansiterm.CreateAnsiTerminal(conn)
	term.Cp437toUtf8 = cp437ToUtf8
	currentSession := session.CreateSession(term, config.TCPRaw, conn.RemoteAddr().String())

	session.Start(currentSession, hangupChannel)
}
