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
	"fmt"
	"io/ioutil"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ConnectionType int

const (
	TCPTelnet ConnectionType = iota
	TCPRaw
	TCPSSH
)

func (t ConnectionType) String() (result string) {
	switch t {
	case TCPTelnet:
		result = "telnet"

	case TCPRaw:
		result = "raw"

	case TCPSSH:
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
	Prometheus struct {
		Enabled bool
		Address string
		Port    uint16
		Path    string
	}
}

// final structure for program options
type ProgramOptions struct {
	LogLevel      log.Level
	SSHPrivateKey string
	Prometheus    PrometheusConfig
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

// final structure for prometheus endpoint
type PrometheusConfig struct {
	Enabled bool
	Address string
	Port    uint16
	Path    string
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
		return nil, fmt.Errorf("Invalid value for log-level. Valid values are: info, error, debug, trace. Received value: %s", config.Options.LogLevel)
	}

	// set default Prometheus values
	AppOptions.Prometheus.Enabled = config.Prometheus.Enabled
	if config.Prometheus.Address == "" {
		AppOptions.Prometheus.Address = "127.0.0.1"
	} else {
		AppOptions.Prometheus.Address = config.Prometheus.Address
	}
	if config.Prometheus.Port == 0 {
		AppOptions.Prometheus.Port = 9000
	} else {
		AppOptions.Prometheus.Port = config.Prometheus.Port
	}
	if config.Prometheus.Path == "" {
		AppOptions.Prometheus.Path = "/metrics"
	} else {
		AppOptions.Prometheus.Path = config.Prometheus.Path
	}

	// set private key for ssh listeners
	AppOptions.SSHPrivateKey = config.Options.SSHPrivateKey
	// validate listener configuration
	if len(config.Listeners) == 0 {
		return nil, fmt.Errorf("No listeners are defined in the configuration file")
	}
	listeners := []Listener{}
	for _, cfgListener := range config.Listeners {
		switch strings.ToLower(cfgListener.Protocol) {
		case "telnet":
			l := Listener{
				Address:     cfgListener.Address,
				Port:        cfgListener.Port,
				ListenType:  TCPTelnet,
				ConvertUTF8: cfgListener.ConvertUTF8,
			}
			listeners = append(listeners, l)
		case "ssh":
			l := Listener{
				Address:     cfgListener.Address,
				Port:        cfgListener.Port,
				ListenType:  TCPSSH,
				ConvertUTF8: cfgListener.ConvertUTF8,
			}
			listeners = append(listeners, l)

		case "raw":
			l := Listener{
				Address:     cfgListener.Address,
				Port:        cfgListener.Port,
				ListenType:  TCPRaw,
				ConvertUTF8: cfgListener.ConvertUTF8,
			}
			listeners = append(listeners, l)

		default:
			return nil, fmt.Errorf("Invalid value for protocol. Valid values are: ssh, telnet, raw. Received value: %s", cfgListener.Protocol)
		}
	}
	return listeners, nil
}
