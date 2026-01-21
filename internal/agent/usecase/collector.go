package usecase

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type Collector struct {
	report    MetricsReporter
	pollCount int64
	mu        sync.Mutex
	metrics   map[string]domain.Metric
}

func NewCollector(report MetricsReporter) *Collector {
	return &Collector{
		report:  report,
		metrics: make(map[string]domain.Metric),
	}
}

func (c *Collector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	c.pollCount++

	gaugeMetrics := map[string]float64{
		"Alloc":         float64(mem.Alloc),
		"BuckHashSys":   float64(mem.BuckHashSys),
		"Frees":         float64(mem.Frees),
		"GCCPUFraction": mem.GCCPUFraction,
		"GCSys":         float64(mem.GCSys),
		"HeapAlloc":     float64(mem.HeapAlloc),
		"HeapIdle":      float64(mem.HeapIdle),
		"HeapInuse":     float64(mem.HeapInuse),
		"HeapObjects":   float64(mem.HeapObjects),
		"HeapReleased":  float64(mem.HeapReleased),
		"HeapSys":       float64(mem.HeapSys),
		"LastGC":        float64(mem.LastGC),
		"Lookups":       float64(mem.Lookups),
		"MCacheInuse":   float64(mem.MCacheInuse),
		"MCacheSys":     float64(mem.MCacheSys),
		"MSpanInuse":    float64(mem.MSpanInuse),
		"MSpanSys":      float64(mem.MSpanSys),
		"Mallocs":       float64(mem.Mallocs),
		"NextGC":        float64(mem.NextGC),
		"NumForcedGC":   float64(mem.NumForcedGC),
		"NumGC":         float64(mem.NumGC),
		"OtherSys":      float64(mem.OtherSys),
		"PauseTotalNs":  float64(mem.PauseTotalNs),
		"StackInuse":    float64(mem.StackInuse),
		"StackSys":      float64(mem.StackSys),
		"Sys":           float64(mem.Sys),
		"TotalAlloc":    float64(mem.TotalAlloc),
		"RandomValue":   rand.Float64(),
	}

	for name, value := range gaugeMetrics {
		c.metrics[name] = domain.Metric{
			Name:  name,
			Type:  domain.Gauge,
			Value: value,
		}
	}

	c.metrics["PollCount"] = domain.Metric{
		Name:  "PollCount",
		Type:  domain.Counter,
		Value: c.pollCount,
	}
}

func (c *Collector) CollectSystemMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	vmem, err := mem.VirtualMemory()
	if err == nil {
		c.metrics["TotalMemory"] = domain.Metric{
			Name:  "TotalMemory",
			Type:  domain.Gauge,
			Value: float64(vmem.Total),
		}

		c.metrics["FreeMemory"] = domain.Metric{
			Name:  "FreeMemory",
			Type:  domain.Gauge,
			Value: float64(vmem.Free),
		}
	}

	cpuPercentages, err := cpu.Percent(0, true)
	if err == nil {
		for i, percentage := range cpuPercentages {
			metricName := fmt.Sprintf("CPUutilization%d", i+1)
			c.metrics[metricName] = domain.Metric{
				Name:  metricName,
				Type:  domain.Gauge,
				Value: percentage,
			}
		}
	}
}

func (c *Collector) GetMetrics() []domain.Metric {
	c.mu.Lock()
	defer c.mu.Unlock()
	metricList := make([]domain.Metric, 0, len(c.metrics))
	for _, metric := range c.metrics {
		metricList = append(metricList, metric)
	}
	return metricList
}

func (c *Collector) Report() {
	c.mu.Lock()
	defer c.mu.Unlock()
	metricList := make([]domain.Metric, 0, len(c.metrics))
	for _, metric := range c.metrics {
		metricList = append(metricList, metric)
	}
	c.report.MetricsReport(metricList)
}

type MetricsReporter interface {
	MetricsReport(metrics []domain.Metric)
}
