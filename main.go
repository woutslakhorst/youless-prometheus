package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// YoulessCollector implements the Collector interface
//[{"tm":1575316361,"net": 1133.932,"pwr": 431,"ts0":1535271600,"cs0": 0.000,"ps0": 0,"p1": 4590.448,"p2": 4315.399,"n1": 2320.876,"n2": 5451.039,"gas": 2878.709,"gts":1912022000}]
type YoulessCollector struct {
	increment uint
	url       string
}

// YoulessEntry is the json format returned from the Youless device
type YoulessEntry struct {
	Tm  uint    // unix-time-format (1489333828 => Sun, 12 Mar 2017 15:50:28 GMT)
	Net float64 // Netto counter, as displayed in the web-interface of the LS-120. It seems equal to: p1 + p2 - n1 - n2 Perhaps also includes some user set offset.
	Pwr int     // Actual power use in Watt (can be negative)
	Ts0 uint    // S0: Unix timestamp of the last S0 measurement.
	Cs0 float64 // S0: kWh counter of S0 input
	Ps0 uint    // S0: Computed power
	P1  float64 // P1 consumption counter (low tariff)
	P2  float64 // P2 consumption counter (high tariff)
	N1  float64 // N1 production counter (low tariff)
	N2  float64 // N2 production counter (high tariff)
	Gas float64 // counter gas-meter (in m^3)
	Gts uint    // Last timestamp created by gas meter yyMMddhhmm
}

// Descriptors used by the YoulessCollector.
var (
	consCountDesc = prometheus.NewDesc(
		"youless_kwh_consumption_total",
		"Kwh power consumption.",
		[]string{"tariff"}, nil,
	)
	prodCountDesc = prometheus.NewDesc(
		"youless_kwh_production_total",
		"Kwh power production.",
		[]string{"tariff"}, nil,
	)
	gasCountDesc = prometheus.NewDesc(
		"youless_gas_consumption_total",
		"gas consumption.",
		[]string{}, nil,
	)
	pwrGaugeDesc = prometheus.NewDesc(
		"youless_pwr_current",
		"Current watt consumption",
		[]string{}, nil,
	)
)

// Call youless URL, fakes for now
func (yc *YoulessCollector) RemoteMetrics() YoulessEntry {
	resp, err := http.Get(yc.url)

	if err != nil {
		log.Printf("Failed calling youless: %v", err)
		return YoulessEntry{}
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed reading response from youless: %v", err)
		return YoulessEntry{}
	}

	var entries []YoulessEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Printf("Failed unmarshalling response from youless: %v", err)
		return YoulessEntry{}
	}

	return entries[0]
}

// Describe is implemented with DescribeByCollect. That's possible because the
// Collect method will always return the same metrics with the same descriptors.
func (yc *YoulessCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(yc, ch)
}

// Collect first triggers the collection of metrics at the youless URL.
func (yc *YoulessCollector) Collect(ch chan<- prometheus.Metric) {
	measurements := yc.RemoteMetrics()

	// Gas
	ch <- prometheus.MustNewConstMetric(gasCountDesc, prometheus.CounterValue, measurements.Gas)

	// P1
	ch <- prometheus.MustNewConstMetric(consCountDesc, prometheus.CounterValue, measurements.P1, "low")

	// P2
	ch <- prometheus.MustNewConstMetric(consCountDesc, prometheus.CounterValue, measurements.P2, "high")

	// N1
	ch <- prometheus.MustNewConstMetric(prodCountDesc, prometheus.CounterValue, measurements.N1, "low")

	// N2
	ch <- prometheus.MustNewConstMetric(prodCountDesc, prometheus.CounterValue, measurements.N2, "high")

	// pwr
	ch <- prometheus.MustNewConstMetric(pwrGaugeDesc, prometheus.GaugeValue, float64(measurements.Pwr))
}

func main() {
	reg := prometheus.NewRegistry()

	// Add the Youless collector.
	reg.MustRegister(
		&YoulessCollector{url: "http://youless/e"},
	)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
