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

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ConnectionType int

const (
	TCP_TELNET ConnectionType = iota
	TCP_RAW
	TCP_SSH
)

func (t ConnectionType) String() (result string) {
	switch t {
	case TCP_TELNET:
		result = "telnet"

	case TCP_RAW:
		result = "raw"

	case TCP_SSH:
		result = "ssh"
	default:
		result = "unknown"
	}
	return
}

// structure used for the yaml parsing
type Config struct {
	Database struct {
		Host     string
		Name     string
		User     string
		Password string
		Port     int
	}
	Options struct {
		LogLevel      string `yaml:"logLevel"`
		SSHPrivateKey string `yaml:"sshPrivateKey"`
	}

	Listeners []struct {
		Address     string
		Port        uint16
		Protocol    string
		ConvertUTF8 bool `yaml:"convertUTF8"`
	}
}

// final structure for program options
type ProgramOptions struct {
	LogLevel      log.Level
	SSHPrivateKey string
}

// final structure for listener config
type Listener struct {
	Address     string
	Port        uint16
	ListenType  ConnectionType
	ConvertUTF8 bool
}

// final structure for db config
type DatabaseConfig struct {
	Host     string
	Port     uint16
	Database string
	User     string
	Password string
}

// package variables for config
var (
	AppOptions = ProgramOptions{
		LogLevel: log.InfoLevel,
	}
)

func ParseConfig(configFile string) ([]Listener, error) {
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}
	log.Trace(config)

	// validate program options
	switch strings.ToLower(config.Options.LogLevel) {
	case "info":
		AppOptions.LogLevel = log.InfoLevel
	case "debug":
		AppOptions.LogLevel = log.DebugLevel
	case "trace":
		AppOptions.LogLevel = log.TraceLevel
	case "error":
		AppOptions.LogLevel = log.ErrorLevel
	default:
		// Unknown level, generate error
		return nil, errors.New(fmt.Sprintf("Invalid value for log-level. Valid values are: info, error, debug, trace. Received value: %s", config.Options.LogLevel))
	}
	// set private key for ssh listeners
	AppOptions.SSHPrivateKey = config.Options.SSHPrivateKey
	// validate listener configuration
	if len(config.Listeners) == 0 {
		return nil, errors.New("No listeners are defined in the configuration file.")
	}
	listeners := []Listener{}
	for _, cfgListener := range config.Listeners {
		switch strings.ToLower(cfgListener.Protocol) {
		case "telnet":
			l := Listener{
				Address:     cfgListener.Address,
				Port:        cfgListener.Port,
				ListenType:  TCP_TELNET,
				ConvertUTF8: cfgListener.ConvertUTF8,
			}
			listeners = append(listeners, l)
		case "ssh":
			l := Listener{
				Address:     cfgListener.Address,
				Port:        cfgListener.Port,
				ListenType:  TCP_SSH,
				ConvertUTF8: cfgListener.ConvertUTF8,
			}
			listeners = append(listeners, l)

		case "raw":
			l := Listener{
				Address:     cfgListener.Address,
				Port:        cfgListener.Port,
				ListenType:  TCP_RAW,
				ConvertUTF8: cfgListener.ConvertUTF8,
			}
			listeners = append(listeners, l)

		default:
			return nil, errors.New(fmt.Sprintf("Invalid value for protocol. Valid values are: ssh, telnet, raw. Received value: %s", cfgListener.Protocol))
		}
	}
	return listeners, nil
}
