package usecase

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type Collector struct {
	sender    MetricSender
	pollCount int64
	mu        sync.Mutex
	metrics   map[string]domain.Metric
}

func NewCollector(sender MetricSender) *Collector {
	return &Collector{
		sender:  sender,
		metrics: make(map[string]domain.Metric),
	}
}

func (c *Collector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	c.pollCount++

	c.metrics["Alloc"] = domain.Metric{Name: "Alloc", Type: domain.Gauge, Value: float64(mem.Alloc)}
	c.metrics["BuckHashSys"] = domain.Metric{Name: "BuckHashSys", Type: domain.Gauge, Value: float64(mem.BuckHashSys)}
	c.metrics["Frees"] = domain.Metric{Name: "Frees", Type: domain.Gauge, Value: float64(mem.Frees)}
	c.metrics["GCCPUFraction"] = domain.Metric{Name: "GCCPUFraction", Type: domain.Gauge, Value: mem.GCCPUFraction}
	c.metrics["GCSys"] = domain.Metric{Name: "GCSys", Type: domain.Gauge, Value: float64(mem.GCSys)}
	c.metrics["HeapAlloc"] = domain.Metric{Name: "HeapAlloc", Type: domain.Gauge, Value: float64(mem.HeapAlloc)}
	c.metrics["HeapIdle"] = domain.Metric{Name: "HeapIdle", Type: domain.Gauge, Value: float64(mem.HeapIdle)}
	c.metrics["HeapInuse"] = domain.Metric{Name: "HeapInuse", Type: domain.Gauge, Value: float64(mem.HeapInuse)}
	c.metrics["HeapObjects"] = domain.Metric{Name: "HeapObjects", Type: domain.Gauge, Value: float64(mem.HeapObjects)}
	c.metrics["HeapReleased"] = domain.Metric{Name: "HeapReleased", Type: domain.Gauge, Value: float64(mem.HeapReleased)}
	c.metrics["HeapSys"] = domain.Metric{Name: "HeapSys", Type: domain.Gauge, Value: float64(mem.HeapSys)}
	c.metrics["LastGC"] = domain.Metric{Name: "LastGC", Type: domain.Gauge, Value: float64(mem.LastGC)}
	c.metrics["Lookups"] = domain.Metric{Name: "Lookups", Type: domain.Gauge, Value: float64(mem.Lookups)}
	c.metrics["MCacheInuse"] = domain.Metric{Name: "MCacheInuse", Type: domain.Gauge, Value: float64(mem.MCacheInuse)}
	c.metrics["MCacheSys"] = domain.Metric{Name: "MCacheSys", Type: domain.Gauge, Value: float64(mem.MCacheSys)}
	c.metrics["MSpanInuse"] = domain.Metric{Name: "MSpanInuse", Type: domain.Gauge, Value: float64(mem.MSpanInuse)}
	c.metrics["MSpanSys"] = domain.Metric{Name: "MSpanSys", Type: domain.Gauge, Value: float64(mem.MSpanSys)}
	c.metrics["Mallocs"] = domain.Metric{Name: "Mallocs", Type: domain.Gauge, Value: float64(mem.Mallocs)}
	c.metrics["NextGC"] = domain.Metric{Name: "NextGC", Type: domain.Gauge, Value: float64(mem.NextGC)}
	c.metrics["NumForcedGC"] = domain.Metric{Name: "NumForcedGC", Type: domain.Gauge, Value: float64(mem.NumForcedGC)}
	c.metrics["NumGC"] = domain.Metric{Name: "NumGC", Type: domain.Gauge, Value: float64(mem.NumGC)}
	c.metrics["OtherSys"] = domain.Metric{Name: "OtherSys", Type: domain.Gauge, Value: float64(mem.OtherSys)}
	c.metrics["PauseTotalNs"] = domain.Metric{Name: "PauseTotalNs", Type: domain.Gauge, Value: float64(mem.PauseTotalNs)}
	c.metrics["StackInuse"] = domain.Metric{Name: "StackInuse", Type: domain.Gauge, Value: float64(mem.StackInuse)}
	c.metrics["StackSys"] = domain.Metric{Name: "StackSys", Type: domain.Gauge, Value: float64(mem.StackSys)}
	c.metrics["Sys"] = domain.Metric{Name: "Sys", Type: domain.Gauge, Value: float64(mem.Sys)}
	c.metrics["TotalAlloc"] = domain.Metric{Name: "TotalAlloc", Type: domain.Gauge, Value: float64(mem.TotalAlloc)}

	c.metrics["PollCount"] = domain.Metric{Name: "PollCount", Type: domain.Counter, Value: c.pollCount}
	c.metrics["RandomValue"] = domain.Metric{Name: "RandomValue", Type: domain.Gauge, Value: rand.Float64()}
}

func (c *Collector) Report() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, metric := range c.metrics {
		err := c.sender.Send(metric)
		if err != nil {
			fmt.Printf("error sending metric: %v\n", err)
		}
	}
}
