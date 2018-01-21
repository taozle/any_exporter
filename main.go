package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PM25 struct {
	AQI              int    `json:"aqi"`
	Area             string `json:"area"`
	PM25             int    `json:"pm2_5"`
	PM25In24H        int    `json:"pm2_5_24h"`
	Position         string `json:"position_name"`
	PrimaryPollutant string `json:"primary_pollutant"`
	Quality          string `json:"quality"`
	StationCode      string `json:"station_code"`
}

var (
	pm25Vec = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "airquality",
			Subsystem: "beijing",
			Name:      "pm25",
			Help:      "Air quality distributions.",
		},
		[]string{"district"},
	)
)

func init() {
	prometheus.MustRegister(pm25Vec)
}

func main() {
	go collectPM25()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":12345", nil))
}

func collectPM25() {
	for range time.Tick(30 * time.Second) {
		for _, v := range request() {
			if v.Position != "" {
				pm25Vec.WithLabelValues(v.Position).Set(float64(v.PM25))
			}
		}
	}
}

func request() []*PM25 {
	url := "http://www.pm25.in/api/querys/pm2_5.json?city=beijing&token=5j1znBVAsnSf5xQyNQyq"

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Retrieving air quality failed, err: %v\n", err)
		return nil
	}

	var v []*PM25
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		log.Printf("Decode resp failed, err: %v\n", err)
		return nil
	}

	return v
}
