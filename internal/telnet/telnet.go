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

package telnet

import (
	"bytes"
	"encoding/binary"
	"net"

	log "github.com/sirupsen/logrus"
)

type ConnectionState int

const (
	// connection state, used during read operations to process incoming telnet commands

	stateData ConnectionState = iota
	stateCR
	stateCommand
	stateWill
	stateWont
	stateDo
	stateDont
	stateSubNeg
	stateSubnegIAC
)

const (
	chCR  byte = 13
	chLF  byte = 10
	chNUL byte = 0

	IAC byte = 255 // Interpret As Command

	// telnet commands

	cmdSE        byte = 240 // End sub negotiation (240)
	cmdNop       byte = 241 // No operation (241)
	cmdData      byte = 242 // Data mark (242)
	cmdBreak     byte = 243 // Break (243)
	cmdIP        byte = 244 // Interrupt process (244)
	cmdAbort     byte = 245 // Abort output (245)
	cmdAYT       byte = 246 // Are you there (246)
	cmdEraseChar byte = 247 // Erase character (247)
	cmdEraseLine byte = 248 // Erase line (248)
	cmdGA        byte = 249 // Go ahead (249)
	cmdSB        byte = 250 // Start sub negotiation (250)
	cmdWill      byte = 251
	cmdWont      byte = 252
	cmdDo        byte = 253
	cmdDont      byte = 254

	// telnet options

	optBinary     byte = 0
	optEcho       byte = 1
	optSupressGA  byte = 3
	optMsgSizeNeg byte = 4
	optStatus     byte = 5
	optTimingMark byte = 6
	optNAWS       byte = 31
)

// Telnet specific connection stuff

type Conn struct {
	net.Conn
	// used during telnet command processing. See ConnectionState constants above.
	// We need to store the state across multiple reads. This seemed like a good spot.
	readState ConnectionState
	// Bytes received for sub negotiation. We also need to store this across multiple reads.
	subNegBuffer bytes.Buffer
	// term info we received
	resizeHandler func(int, int)
}

func (c *Conn) SendCommand(cmd byte) error {
	buffer := []byte{
		IAC,
		cmd,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendWill(o byte) error {
	const cmd byte = 251
	log.Debugf("%s - Send WILL: %d", c.RemoteAddr(), o)
	buffer := []byte{
		IAC,
		cmd,
		o,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendWont(o byte) error {
	const cmd byte = 252
	log.Debugf("%s - Send WONT: %d", c.RemoteAddr(), o)
	buffer := []byte{
		IAC,
		cmd,
		o,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendDo(o byte) error {
	const cmd byte = 253
	log.Debugf("%s - Send DO: %d", c.RemoteAddr(), o)
	buffer := []byte{
		IAC,
		cmd,
		o,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendDont(o byte) error {
	const cmd byte = 254
	log.Debugf("%s - Send DONT: %d", c.RemoteAddr(), o)
	buffer := []byte{
		IAC,
		cmd,
		o,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

// let's make it compatible with a standard io writer.

// I have no idea if this code works ok under all circumstances (sudden disconnects etc..). Only time will tell.
// IAC needs to be escaped (=duplicated), otherwise telnet client thinks we are sending a telnet command.
func (c *Conn) Write(data []byte) (totalWritten int, err error) {
	for len(data) > 0 {
		var currentWritten int
		index := bytes.IndexByte(data, IAC)
		if index == -1 {
			currentWritten, err = c.Conn.Write(data)
			totalWritten += currentWritten
			break
		} else {
			// write everything before the IAC byte
			currentWritten, err = c.Conn.Write(data[:index])
			totalWritten += currentWritten
			if err != nil {
				log.Errorln(err.Error())
				break
			}
			// write double IAC for escaping purposes
			currentWritten, err = c.Conn.Write([]byte{IAC, IAC})
			// not sure if we should account for the fact that more data is written than the original buffer
			// (because of the IAC doubling). TO-DO: Investigate later.
			totalWritten += currentWritten
			if err != nil {
				log.Errorln(err.Error())
				break
			}
			// Repeat all this for the rest of the buffer
			data = data[index+1:]
		}
	}
	return
}

// Here we read data, and process any telnet command if they are encountered.
// Escaped IAC bytes (= double IAC byte) should be unescaped here as well.
func (c *Conn) Read(data []byte) (int, error) {
	destIndex := 0
	buffer := make([]byte, len(data)) // make a new buffer for reading data, same size as original one
	tempRead, err := c.Conn.Read(buffer)
	// process all bytes that have been read, even if an error occurred.
	// This seems to be a recommended approach:
	//
	// https://golang.org/pkg/io/
	//
	// Callers should always process the n > 0 bytes returned before considering the error err.
	// Doing so correctly handles I/O errors that happen after reading some bytes and also both of the allowed EOF behaviors.
	//
	if tempRead > 0 {
		for _, element := range buffer[:tempRead] {
			switch c.readState {
			case stateData:
				switch element {
				case IAC:
					log.Tracef("%s - State changed to stateCommand", c.RemoteAddr())
					c.readState = stateCommand
				case chCR:
					data[destIndex] = element
					destIndex++
					log.Tracef("%s - State changed to stateCR", c.RemoteAddr())
					c.readState = stateCR
				default:
					// not IAC or CR, so it's data
					data[destIndex] = element
					destIndex++
				}

			case stateCR:
				// only for conversion from CR/LF or CR/NUL to CR. We probably shouldn't do this if telnet is in binary mode.
				// TO-DO implement binary mode support :-D
				switch element {
				case chLF, chNUL:
					// Do nothing. just ignore this byte.
				default:
					// Add to buffer, as it's just normal data
					data[destIndex] = element
					destIndex++

				}
				// switch back to normal data state
				log.Tracef("%s - State changed to stateData", c.RemoteAddr())
				c.readState = stateData

			case stateCommand:
				if element == IAC {
					// this is en escaped byte 255, so just consider it as data
					data[destIndex] = element
					destIndex++
					log.Tracef("%s - State changed to stateData", c.RemoteAddr())
					c.readState = stateData
				} else {
					// it's a telnet command. Set our processor to the correct state before processing the next byte
					switch element {
					case cmdWill:
						log.Tracef("%s - State changed to stateWill", c.RemoteAddr())
						c.readState = stateWill
					case cmdDo:
						log.Tracef("%s - State changed to stateDo", c.RemoteAddr())
						c.readState = stateDo
					case cmdWont:
						log.Tracef("%s - State changed to stateWont", c.RemoteAddr())
						c.readState = stateWont
					case cmdDont:
						log.Tracef("%s - State changed to stateDont", c.RemoteAddr())
						c.readState = stateDont
					case cmdSB:
						log.Tracef("%s - State changed to stateSubNeg", c.RemoteAddr())
						c.readState = stateSubNeg
						// start of a new subnegotation, so let's reset the buffer
						c.subNegBuffer.Reset()
					default:
						log.Debugf("%s - Received telnet command: %d", c.RemoteAddr(), element)
						c.commandHandler(element)
						log.Tracef("%s - State changed to stateData", c.RemoteAddr())
						c.readState = stateData
					}
				}
			case stateWill:
				log.Debugf("%s - Received WILL: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to stateData", c.RemoteAddr())
				c.optionHandler(cmdWill, element)
				c.readState = stateData

			case stateWont:
				log.Debugf("%s - Received WONT: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to stateData", c.RemoteAddr())
				c.optionHandler(cmdWont, element)
				c.readState = stateData

			case stateDo:
				log.Debugf("%s - Received DO: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to stateData", c.RemoteAddr())
				c.optionHandler(cmdDo, element)
				c.readState = stateData

			case stateDont:
				log.Debugf("%s - Received DONT: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to stateData", c.RemoteAddr())
				c.optionHandler(cmdDont, element)
				c.readState = stateData

			// here we do all the subnegotation handling. This is a bit messy...

			case stateSubNeg:
				switch element {
				case IAC:
					log.Tracef("%s - State changed to stateSubnegIAC", c.RemoteAddr())
					c.readState = stateSubnegIAC
				default:
					// default action is to add data to the buffer
					c.subNegBuffer.Grow(1)
					err := c.subNegBuffer.WriteByte(element)
					if err != nil {
						log.Errorln(err.Error())
					}

				}
			case stateSubnegIAC:
				switch element {
				case IAC:
					// it was an escaped IAC, so add it to the subneg buffer
					c.subNegBuffer.Grow(1)
					err := c.subNegBuffer.WriteByte(element)
					if err != nil {
						log.Errorln(err.Error())
					}
					log.Tracef("%s - State changed to stateSubNeg", c.RemoteAddr())
					c.readState = stateSubNeg

				case cmdSE:
					// received end of subnegotation. Call handler and move back to data mode
					c.subNegHandler()
					log.Tracef("%s - State changed to stateData", c.RemoteAddr())
					c.readState = stateData
				}

			}
		}

	}
	return destIndex, err
}

func (c *Conn) RequestTermSize() {
	err := c.SendDo(optNAWS)
	if err != nil {
		log.Errorln(err.Error())
	}
}

// create and initialize telnet connection object
func NewConnection(c net.Conn) *Conn {
	conn := Conn{
		Conn: c,
	}
	// set telnet parameters, this should ensure the connection is in character-mode, and echo'ing is done by the server.
	// We want to be in control of all echo-ing.
	err := conn.SendWill(optSupressGA)
	if err != nil {
		log.Errorln(err.Error())
	}
	err = conn.SendWill(optEcho)
	if err != nil {
		log.Errorln(err.Error())
	}

	return &conn
}

// private functions

func (c *Conn) subNegHandler() {
	data := c.subNegBuffer.Bytes()
	log.Traceln(data)

	if len(data) > 0 {
		option := data[0]
		switch option {
		case optNAWS:
			if len(data) != 5 {
				log.Errorf("%s - Incorrect amount of parameters for NAWS subnegotiation.", c.RemoteAddr())
				break
			}
			w := binary.BigEndian.Uint16(data[1:3])
			h := binary.BigEndian.Uint16(data[3:5])
			// only update if size is bigger than zero.
			// call resizeHandler if installed
			if c.resizeHandler != nil {
				c.resizeHandler(int(w), int(h))
			}
			log.Debugf("%s - terminal size update received (w=%d, h=%d)", c.RemoteAddr(), w, h)
		default:
			log.Debugf("%s - Unknown subnegotation received (%d). Ignoring.", c.RemoteAddr(), option)
		}
	}
	// always reset the buffer after we handled the subnegotiation
	c.subNegBuffer.Reset()

}

func (c *Conn) InstallResizeHandler(handler func(int, int)) {
	c.resizeHandler = handler

}

func (c *Conn) commandHandler(command byte) {
	// TO-DO: implement if necessary
}

// I doubt this is correct behaviour. I should investigate the Q Method (RFC1143) for option negotiation.
func (c *Conn) optionHandler(command byte, option byte) {
	var err error
	switch command {
	case cmdWill:
		switch option {
		// only send DO for options we actually support
		case optNAWS, optSupressGA:
			err = c.SendDo(option)
		case optEcho:
			// explicitly disable local echo on client for now? Should this be allowed if client requests it?
			err = c.SendDont(option)
		default:
			err = c.SendDont(option)
		}
	case cmdWont:
	case cmdDo:
		switch option {
		case optSupressGA, optEcho:
			err = c.SendWill(option)
		default:
			err = c.SendWont(option)

		}
	case cmdDont:
	}
	if err != nil {
		log.Errorln(err.Error())
	}
}
