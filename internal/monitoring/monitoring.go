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

package monitoring

import (
	"fmt"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/jeroenjacobs79/tobw/internal/config"
	log "github.com/sirupsen/logrus"
)

func StartMetricsEndpoint(config config.PrometheusConfig) {
	log.Infof("Starting Prometheus metrics listener on address: http://%s:%d%s...", config.Address, config.Port, config.Path)
	http.Handle(config.Path, promhttp.Handler())
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.Address, config.Port), nil)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Infof("Prometheus metrics listener started.")
}
