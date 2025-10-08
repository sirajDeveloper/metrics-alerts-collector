package usecase

import (
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type Collector struct {
	sender    MetricSender
	pollCount int64
	metrics   map[string]domain.Metric
}

func NewCollector(sender MetricSender) *Collector {
	return &Collector{
		sender:  sender,
		metrics: make(map[string]domain.Metric),
	}
}

func (c *Collector) Collect() {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	c.pollCount++

	c.metrics["Alloc"] = domain.Metric{"Alloc", domain.Gauge, float64(mem.Alloc)}
	c.metrics["BuckHashSys"] = domain.Metric{"BuckHashSys", domain.Gauge, float64(mem.BuckHashSys)}
	c.metrics["Frees"] = domain.Metric{"Frees", domain.Gauge, float64(mem.Frees)}
	c.metrics["GCCPUFraction"] = domain.Metric{"GCCPUFraction", domain.Gauge, mem.GCCPUFraction}
	c.metrics["GCSys"] = domain.Metric{"GCSys", domain.Gauge, float64(mem.GCSys)}
	c.metrics["HeapAlloc"] = domain.Metric{"HeapAlloc", domain.Gauge, float64(mem.HeapAlloc)}
	c.metrics["HeapIdle"] = domain.Metric{"HeapIdle", domain.Gauge, float64(mem.HeapIdle)}
	c.metrics["HeapInuse"] = domain.Metric{"HeapInuse", domain.Gauge, float64(mem.HeapInuse)}
	c.metrics["HeapObjects"] = domain.Metric{"HeapObjects", domain.Gauge, float64(mem.HeapObjects)}
	c.metrics["HeapReleased"] = domain.Metric{"HeapReleased", domain.Gauge, float64(mem.HeapReleased)}
	c.metrics["HeapSys"] = domain.Metric{"HeapSys", domain.Gauge, float64(mem.HeapSys)}
	c.metrics["LastGC"] = domain.Metric{"LastGC", domain.Gauge, float64(mem.LastGC)}
	c.metrics["Lookups"] = domain.Metric{"Lookups", domain.Gauge, float64(mem.Lookups)}
	c.metrics["MCacheInuse"] = domain.Metric{"MCacheInuse", domain.Gauge, float64(mem.MCacheInuse)}
	c.metrics["MCacheSys"] = domain.Metric{"MCacheSys", domain.Gauge, float64(mem.MCacheSys)}
	c.metrics["MSpanInuse"] = domain.Metric{"MSpanInuse", domain.Gauge, float64(mem.MSpanInuse)}
	c.metrics["MSpanSys"] = domain.Metric{"MSpanSys", domain.Gauge, float64(mem.MSpanSys)}
	c.metrics["Mallocs"] = domain.Metric{"Mallocs", domain.Gauge, float64(mem.Mallocs)}
	c.metrics["NextGC"] = domain.Metric{"NextGC", domain.Gauge, float64(mem.NextGC)}
	c.metrics["NumForcedGC"] = domain.Metric{"NumForcedGC", domain.Gauge, float64(mem.NumForcedGC)}
	c.metrics["NumGC"] = domain.Metric{"NumGC", domain.Gauge, float64(mem.NumGC)}
	c.metrics["OtherSys"] = domain.Metric{"OtherSys", domain.Gauge, float64(mem.OtherSys)}
	c.metrics["PauseTotalNs"] = domain.Metric{"PauseTotalNs", domain.Gauge, float64(mem.PauseTotalNs)}
	c.metrics["StackInuse"] = domain.Metric{"StackInuse", domain.Gauge, float64(mem.StackInuse)}
	c.metrics["StackSys"] = domain.Metric{"StackSys", domain.Gauge, float64(mem.StackSys)}
	c.metrics["Sys"] = domain.Metric{"Sys", domain.Gauge, float64(mem.Sys)}
	c.metrics["TotalAlloc"] = domain.Metric{"TotalAlloc", domain.Gauge, float64(mem.TotalAlloc)}

	c.metrics["PollCount"] = domain.Metric{"PollCount", domain.Counter, c.pollCount}
	c.metrics["RandomValue"] = domain.Metric{"RandomValue", domain.Gauge, rand.Float64()}
}

func (c *Collector) Report() {
	for _, metric := range c.metrics {
		err := c.sender.Send(metric)
		if err != nil {
			fmt.Printf("error sending metric: %v", err)
		}
	}
}
