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

package cfgparse

type ListenerType int

const (
	TCP_RAW ListenerType = 0
	TCP_SSH ListenerType = 1
	TCP_TELNET ListenerType = 2
)

type Listener struct {
	address string
	port uint16
	listenType ListenerType
}


type DatabaseConfig struct {
	host string
	port uint16
	database string
	user string
	password string
}

func ParseConfig(configFile string) (listeners []Listener, dbConfig DatabaseConfig) {
	
}
