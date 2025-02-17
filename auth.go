package main

import (
	"github.com/leejones/netrc"
	"net/http"
	"os"
)

func AddAuth(req *http.Request) {
	apiToken := os.Getenv("GRAFANA_API_TOKEN")
	grafanaUser := os.Getenv("GRAFANA_BASIC_AUTH_USER")
	grafanaPassword := os.Getenv("GRAFANA_BASIC_AUTH_PASSWORD")

	if apiToken != "" {
		req.Header.Add("Authorization", "Bearer "+apiToken)
	} else {
		if grafanaUser == "" || grafanaPassword == "" {
			log.Error("load credentials: ENV vars not set: GRAFANA_BASIC_AUTH_USER, GRAFANA_BASIC_AUTH_PASSWORD")
			basicAuth, err := netrc.Get(req.Host)
			if err != nil {
				log.Errorf("load credentials: unable to load from netrc: %v", err)
			} else {
				log.Info("load credentials: found credentials in netrc")
				grafanaUser = basicAuth.Username
				grafanaPassword = basicAuth.Password
			}
		}
		req.SetBasicAuth(grafanaUser, grafanaPassword)
	}
}
