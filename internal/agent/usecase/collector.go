package usecase

import (
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

	c.metrics["Alloc"] = domain.Metric{"Alloc", Gauge, float64(memStats.Alloc)}
	c.metrics["BuckHashSys"] = domain.Metric{"BuckHashSys", Gauge, float64(memStats.BuckHashSys)}
	c.metrics["Frees"] = domain.Metric{"Frees", Gauge, float64(memStats.Frees)}
	c.metrics["GCCPUFraction"] = domain.Metric{"GCCPUFraction", Gauge, memStats.GCCPUFraction}
	c.metrics["GCSys"] = domain.Metric{"GCSys", Gauge, float64(memStats.GCSys)}
	c.metrics["HeapAlloc"] = domain.Metric{"HeapAlloc", Gauge, float64(memStats.HeapAlloc)}
	c.metrics["HeapIdle"] = domain.Metric{"HeapIdle", Gauge, float64(memStats.HeapIdle)}
	c.metrics["HeapInuse"] = domain.Metric{"HeapInuse", Gauge, float64(memStats.HeapInuse)}
	c.metrics["HeapObjects"] = domain.Metric{"HeapObjects", Gauge, float64(memStats.HeapObjects)}
	c.metrics["HeapReleased"] = domain.Metric{"HeapReleased", Gauge, float64(memStats.HeapReleased)}
	c.metrics["HeapSys"] = domain.Metric{"HeapSys", Gauge, float64(memStats.HeapSys)}
	c.metrics["LastGC"] = domain.Metric{"LastGC", Gauge, float64(memStats.LastGC)}
	c.metrics["Lookups"] = domain.Metric{"Lookups", Gauge, float64(memStats.Lookups)}
	c.metrics["MCacheInuse"] = domain.Metric{"MCacheInuse", Gauge, float64(memStats.MCacheInuse)}
	c.metrics["MCacheSys"] = domain.Metric{"MCacheSys", Gauge, float64(memStats.MCacheSys)}
	c.metrics["MSpanInuse"] = domain.Metric{"MSpanInuse", Gauge, float64(memStats.MSpanInuse)}
	c.metrics["MSpanSys"] = domain.Metric{"MSpanSys", Gauge, float64(memStats.MSpanSys)}
	c.metrics["Mallocs"] = domain.Metric{"Mallocs", Gauge, float64(memStats.Mallocs)}
	c.metrics["NextGC"] = domain.Metric{"NextGC", Gauge, float64(memStats.NextGC)}
	c.metrics["NumForcedGC"] = domain.Metric{"NumForcedGC", Gauge, float64(memStats.NumForcedGC)}
	c.metrics["NumGC"] = domain.Metric{"NumGC", Gauge, float64(memStats.NumGC)}
	c.metrics["OtherSys"] = domain.Metric{"OtherSys", Gauge, float64(memStats.OtherSys)}
	c.metrics["PauseTotalNs"] = domain.Metric{"PauseTotalNs", Gauge, float64(memStats.PauseTotalNs)}
	c.metrics["StackInuse"] = domain.Metric{"StackInuse", Gauge, float64(memStats.StackInuse)}
	c.metrics["StackSys"] = domain.Metric{"StackSys", Gauge, float64(memStats.StackSys)}
	c.metrics["Sys"] = domain.Metric{"Sys", Gauge, float64(memStats.Sys)}
	c.metrics["TotalAlloc"] = domain.Metric{"TotalAlloc", Gauge, float64(memStats.TotalAlloc)}
	c.metrics["RandomValue"] = domain.Metric{"RandomValue", domain.Gauge, rand.Float64() * 100}
}

func (c *Collector) Report() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, metric := range c.metrics {
		_ = c.sender.Send(metric) // Обработка ошибок по необходимости
	}
}
