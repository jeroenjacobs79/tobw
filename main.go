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
	"sync"
	"tobw/internal/termserve"

	log "github.com/sirupsen/logrus"
)

// will be replaced during build-phase with actual git-based version info
var Version = "local"

const (
	APP_NAME = "Tale of the Black Wyvern"
)

func main() {
	// set log format to include timestamp, even when TTY is attached.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true},
	)
	// set log level
	log.SetLevel(log.InfoLevel)
	// startup message
	log.Infof("%s (version %s) is starting up...", APP_NAME, Version)
	// start our listeners
	var wg sync.WaitGroup
	wg.Add(1)
	go termserve.StartListener(&wg, ":5000", termserve.RawTCP, true)

	wg.Add(1)
	go termserve.StartListener(&wg, ":6000", termserve.RawTCP, false)

	wg.Add(1)
	go termserve.StartListener(&wg, ":5023", termserve.Telnet, true)

	wg.Add(1)
	go termserve.StartListener(&wg, ":5022", termserve.Ssh, true)

	wg.Wait()

}
