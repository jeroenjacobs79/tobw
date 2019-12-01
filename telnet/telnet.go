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

package telnet

import (
	"bytes"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"net"
)

type ConnectionState int

const (
	// connection state, used during read operations to process incoming telnet commands

	STATE_DATA ConnectionState = iota
	STATE_CR
	STATE_COMMAND
	STATE_WILL
	STATE_WONT
	STATE_DO
	STATE_DONT
	STATE_SUBNEG
	STATE_SUBNEG_IAC

)

const (

	CH_CR 				byte = 13
	CH_LF 				byte = 10
	CH_NUL 				byte = 0

	IAC 				byte = 255 // Interpret As Command

	// telnet commands

	CMD_SE 				byte = 240 // End sub negotiation (240)
	CMD_NOP 			byte = 241 // No operation (241)
	CMD_DATA 			byte = 242 // Data mark (242)
	CMD_BREAK 			byte = 243 // Break (243)
	CMD_IP 				byte = 244 // Interrupt process (244)
	CMD_ABORT 			byte = 245 // Abort output (245)
	CMD_AYT 			byte = 246 // Are you there (246)
	CMD_ERASE_CHAR 		byte = 247 // Erase character (247)
	CMD_ERASE_LINE 		byte = 248 // Erase line (248)
	CMD_GA 				byte = 249 // Go ahead (249)
	CMD_SB 				byte = 250 // Start sub negotiation (250)
	CMD_WILL			byte = 251
	CMD_WONT			byte = 252
	CMD_DO				byte = 253
	CMD_DONT			byte = 254

	// telnet options

	OPT_BINARY 			byte = 0
	OPT_ECHO 			byte = 1
	OPT_SUPPRESS_GA		byte = 3
	OPT_MSG_SIZE_NEG 	byte = 4
	OPT_STATUS 			byte = 5
	OPT_TIMING_MARK 	byte = 6
	OPT_NAWS			byte = 31
)

// Telnet specific connection stuff

type Conn struct {
	net.Conn
	// used during telnet command processing. See ConnectionState constants above.
	// We need to store the state across multiple reads. This seemed like a good spot.
	readState ConnectionState
	// Bytes received for sub negotation. We also need to store this across multiple reads.
	subNegBuffer bytes.Buffer
	// term info we received
	resizeHandler func(int, int)
}


func (c *Conn) SendCommand(cmd byte) (error) {
	buffer := []byte {
		IAC,
		cmd,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendWill(o byte) (error) {
	const cmd byte = 251
	log.Debugf("%s - Send WILL: %d", c.RemoteAddr(), o)
	buffer := []byte {
		IAC,
		cmd,
		o,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendWont(o byte) (error) {
	const cmd byte = 252
	log.Debugf("%s - Send WONT: %d", c.RemoteAddr(), o)
	buffer := []byte {
		IAC,
		cmd,
		o,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendDo(o byte) (error) {
	const cmd byte = 253
	log.Debugf("%s - Send DO: %d", c.RemoteAddr(), o)
	buffer := []byte {
		IAC,
		cmd,
		o,
	}
	_, err := c.Conn.Write(buffer)
	return err
}

func (c *Conn) SendDont(o byte) (error) {
	const cmd byte = 254
	log.Debugf("%s - Send DONT: %d",c.RemoteAddr(), o)
	buffer := []byte {
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
		if index==-1 {
			currentWritten, err = c.Conn.Write(data)
			totalWritten += currentWritten
			break
		} else {
			// write everything before the IAC byte
			currentWritten, err = c.Conn.Write(data[:index])
			totalWritten += currentWritten
			if err!=nil {
				log.Errorln(err.Error())
				break
			}
			// write double IAC for escaping purposes
			currentWritten, err = c.Conn.Write([]byte {IAC, IAC})
			// not sure if we should account for the fact that more data is written than the original buffer
			// (because of the IAC doubling). TO-DO: Investigate later.
			totalWritten += currentWritten
			if err!=nil {
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
	// process all bytes that have been read, even if an error occured.
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
			case STATE_DATA:
				switch element {
				case IAC:
					log.Tracef("%s - State changed to STATE_COMMAND", c.RemoteAddr())
					c.readState = STATE_COMMAND
				case CH_CR:
					data[destIndex] = element
					destIndex++
					log.Tracef("%s - State changed to STATE_CR", c.RemoteAddr())
					c.readState = STATE_CR
				default:
					// not IAC or CR, so it's data
					data[destIndex] = element
					destIndex++
				}

			case STATE_CR:
				// only for conversion from CR/LF or CR/NUL to CR. We probably shouldn't do this if telnet is in binary mode.
				// TO-DO implement binary mode support :-D
				switch(element) {
				case CH_LF, CH_NUL:
					// Do nothing. just ignore this byte.
				default:
					// Add to buffer, as it's just normal data
					data[destIndex] = element
					destIndex++

				}
				// switch back to normal data state
				log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
				c.readState = STATE_DATA

			case STATE_COMMAND:
				if element==IAC {
					// this is en escaped byte 255, so just consider it as data
					data[destIndex] = element
					destIndex++
					log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
					c.readState = STATE_DATA
				} else {
					// it's a telnet command. Set our processor to the correct state before processing the next byte
					switch element {
					case CMD_WILL:
						log.Tracef("%s - State changed to STATE_WILL", c.RemoteAddr())
						c.readState = STATE_WILL
					case CMD_DO:
						log.Tracef("%s - State changed to STATE_DO", c.RemoteAddr())
						c.readState = STATE_DO
					case CMD_WONT:
						log.Tracef("%s - State changed to STATE_WONT", c.RemoteAddr())
						c.readState = STATE_WONT
					case CMD_DONT:
						log.Tracef("%s - State changed to STATE_DONT", c.RemoteAddr())
						c.readState = STATE_DONT
					case CMD_SB:
						log.Tracef("%s - State changed to STATE_SUBNEG", c.RemoteAddr())
						c.readState = STATE_SUBNEG
						// start of a new subnegotation, so let's reset the buffer
						c.subNegBuffer.Reset()
					default:
						log.Debugf("%s - Received telnet command: %d", c.RemoteAddr(), element)
						c.commandHandler(element)
						log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
						c.readState = STATE_DATA
					}
				}
			case STATE_WILL:
				log.Debugf("%s - Received WILL: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
				c.optionHandler(CMD_WILL, element)
				c.readState = STATE_DATA

			case STATE_WONT:
				log.Debugf("%s - Received WONT: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
				c.optionHandler(CMD_WONT, element)
				c.readState = STATE_DATA

			case STATE_DO:
				log.Debugf("%s - Received DO: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
				c.optionHandler(CMD_DO, element)
				c.readState = STATE_DATA

			case STATE_DONT:
				log.Debugf("%s - Received DONT: %d", c.RemoteAddr(), element)
				log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
				c.optionHandler(CMD_DONT, element)
				c.readState = STATE_DATA

			// here we do all the subnegotation handling. This is a bit messy...

			case STATE_SUBNEG:
				switch element {
				case IAC:
					log.Tracef("%s - State changed to STATE_SUBNEG_IAC", c.RemoteAddr())
					c.readState = STATE_SUBNEG_IAC
				default:
					// default action is to add data to the buffer
					c.subNegBuffer.Grow(1)
					err := c.subNegBuffer.WriteByte(element)
					if err!=nil {
						log.Errorln(err.Error())
					}

				}
			case STATE_SUBNEG_IAC:
				switch element {
				case IAC:
					// it was an escaped IAC, so add it to the subneg buffer
					c.subNegBuffer.Grow(1)
					err := c.subNegBuffer.WriteByte(element)
					if err!=nil {
						log.Errorln(err.Error())
					}
					log.Tracef("%s - State changed to STATE_SUBNEG", c.RemoteAddr())
					c.readState = STATE_SUBNEG

				case CMD_SE:
					// received end of subnegotation. Call handler and move back to data mode
					c.subNegHandler()
					log.Tracef("%s - State changed to STATE_DATA", c.RemoteAddr())
					c.readState = STATE_DATA
				}

			}
		}

	}
	return destIndex, err
}

func (c *Conn) RequestTermSize() {
	err := c.SendDo(OPT_NAWS)
	if err!=nil {
		log.Errorln(err.Error())
	}
}

// create and initialize telnet connection object
func NewConnection(c net.Conn) (*Conn) {
	conn := Conn {
		Conn: c,
	}
	// set telnet parameters, this should ensure the connection is in character-mode, and echo'ing is done by the server.
	// We want to be in control of all echo-ing.
	err := conn.SendWill(OPT_SUPPRESS_GA)
	if err!=nil {
		log.Errorln(err.Error())
	}
	err = conn.SendWill(OPT_ECHO)
	if err!=nil {
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
		case OPT_NAWS:
			if len(data) != 5 {
				log.Errorf("%s - Incorrect amount of parameters for NAWS subnegotiation.", c.RemoteAddr())
				break
			}
			w := binary.BigEndian.Uint16(data[1:3])
			h := binary.BigEndian.Uint16(data[3:5])
			// only update if size is bigger than zero.
			// call resizeHandler if installed
			if c.resizeHandler!=nil {
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
	case CMD_WILL:
		switch(option) {
		// only send DO for options we actually support
		case OPT_NAWS, OPT_SUPPRESS_GA:
			err = c.SendDo(option)
		case OPT_ECHO:
			// explicitly disable local echo on client for now? Should this be allowed if client requests it?
			err = c.SendDont(option)
		default:
			err = c.SendDont(option)
		}
	case CMD_WONT:
	case CMD_DO:
		switch(option) {
		case OPT_SUPPRESS_GA, OPT_ECHO:
			err = c.SendWill(option)
		default:
			err = c.SendWont(option)

		}
	case CMD_DONT:
	}
	if err != nil {
		log.Errorln(err.Error())
	}
}