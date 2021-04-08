package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const fetchInterval = 5 * time.Minute

var (
	statuses = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ap_status",
			Help: "AP connection status",
		},
		[]string{"ap"},
	)
	clients2g = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ap_num_clients_2G",
			Help: "AP clients 2G",
		},
		[]string{"ap"},
	)
	clients5g = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ap_num_clients_5G",
			Help: "AP clients 5G",
		},
		[]string{"ap"},
	)
	client_count = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "client_count",
			Help: "Number of users per SSID",
		},
		[]string{"ssid"},
	)
)

func reset(err error) {
	log.Println("failed to fetch, resetting metrics:", err)
	statuses.Reset()
	clients2g.Reset()
	clients5g.Reset()
	client_count.Reset()
}

func fetch() {
	start := time.Now()
	defer func() {
		log.Println("fetch: completed in:", time.Since(start))
	}()

	rc := NewRuckusClient()
	err := rc.Login()
	if err != nil {
		reset(err)
		return
	}
	aps, err := rc.ListAP()
	if err != nil {
		reset(err)
		return
	}

	for _, ap := range aps.Aps {
		status := statuses.WithLabelValues(ap.Name)
		if ap.Status == "Online" {
			status.Set(1)
		} else {
			status.Set(0)
		}

		clients2g.WithLabelValues(ap.Name).Set(float64(ap.Clients24G))
		clients5g.WithLabelValues(ap.Name).Set(float64(ap.Clients5G))
	}
	wlans, err := rc.WLAN()
	if err != nil {
		reset(err)
		return

	}
	clientBySSID := make(map[string]int)
	for _, ssid := range wlans.WLANs {
		clientBySSID[ssid.Name] += ssid.Clients
	}
	for wlanName, count := range clientBySSID {
		client_count.WithLabelValues(wlanName).Set(float64(count))
	}
}

func main() {
	go func() {
		for {
			fetch()
			time.Sleep(fetchInterval)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9118", nil)
}
