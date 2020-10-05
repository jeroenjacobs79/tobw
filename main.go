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

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jeroenjacobs79/tobw/internal/monitoring"

	"github.com/jeroenjacobs79/tobw/internal/config"
	"github.com/jeroenjacobs79/tobw/internal/termserve"
	log "github.com/sirupsen/logrus"
)

// will be replaced during build-phase with actual git-based version info
var Version = "local"

const (
	AppName = "Tale of the Black Wyvern"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	go cancelOnInterrupt(ctx, cancelFunc)
	err := run(ctx)
	if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		log.Fatalln(err)
	}
}

func cancelOnInterrupt(ctx context.Context, cancelFunction context.CancelFunc) {
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case sig := <-term:
			log.Infof("Received %s, exiting gracefully...", sig)
			cancelFunction()
			os.Exit(0)
		case <-ctx.Done():
			os.Exit(0)
		}
	}
}

func run(ctx context.Context) error {
	// parse commandline for config file. Error if not specified.
	if len(os.Args) != 2 {
		fmt.Printf("%s (version %s)\n", AppName, Version)
		fmt.Println("No config file specified.")
		fmt.Println()
		fmt.Println("Usage:", os.Args[0], "/path/to/config.yaml")
		return nil
	}

	// start parsing config
	listeners, err := config.ParseConfig(os.Args[1])
	if err != nil {
		return err
	}

	// set log format to include timestamp, even when TTY is attached.
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true},
	)
	// set log level
	log.SetLevel(config.AppOptions.LogLevel)
	// startup message
	log.Infof("%s (version %s) is starting up...", AppName, Version)
	// start metrics endpoint, if configured
	if config.AppOptions.Prometheus.Enabled {
		go monitoring.StartMetricsEndpoint(config.AppOptions.Prometheus)
	}
	// start our listeners
	var wg sync.WaitGroup
	for _, listener := range listeners {
		wg.Add(1)
		go termserve.StartListener(&wg, fmt.Sprintf("%s:%d", listener.Address, listener.Port), listener.ListenType, listener.ConvertUTF8)
	}
	wg.Wait()
	return nil
}
