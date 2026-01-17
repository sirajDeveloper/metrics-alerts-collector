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
	report       MetricsReporter
	pollCount    int64
	mu           sync.Mutex
	metrics      map[string]domain.Metric
	gaugeMetrics map[string]float64
}

func NewCollector(report MetricsReporter) *Collector {
	return &Collector{
		report:       report,
		metrics:      make(map[string]domain.Metric, 30),
		gaugeMetrics: make(map[string]float64, 28),
	}
}

func (c *Collector) Collect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	c.pollCount++

	c.gaugeMetrics["Alloc"] = float64(mem.Alloc)
	c.gaugeMetrics["BuckHashSys"] = float64(mem.BuckHashSys)
	c.gaugeMetrics["Frees"] = float64(mem.Frees)
	c.gaugeMetrics["GCCPUFraction"] = mem.GCCPUFraction
	c.gaugeMetrics["GCSys"] = float64(mem.GCSys)
	c.gaugeMetrics["HeapAlloc"] = float64(mem.HeapAlloc)
	c.gaugeMetrics["HeapIdle"] = float64(mem.HeapIdle)
	c.gaugeMetrics["HeapInuse"] = float64(mem.HeapInuse)
	c.gaugeMetrics["HeapObjects"] = float64(mem.HeapObjects)
	c.gaugeMetrics["HeapReleased"] = float64(mem.HeapReleased)
	c.gaugeMetrics["HeapSys"] = float64(mem.HeapSys)
	c.gaugeMetrics["LastGC"] = float64(mem.LastGC)
	c.gaugeMetrics["Lookups"] = float64(mem.Lookups)
	c.gaugeMetrics["MCacheInuse"] = float64(mem.MCacheInuse)
	c.gaugeMetrics["MCacheSys"] = float64(mem.MCacheSys)
	c.gaugeMetrics["MSpanInuse"] = float64(mem.MSpanInuse)
	c.gaugeMetrics["MSpanSys"] = float64(mem.MSpanSys)
	c.gaugeMetrics["Mallocs"] = float64(mem.Mallocs)
	c.gaugeMetrics["NextGC"] = float64(mem.NextGC)
	c.gaugeMetrics["NumForcedGC"] = float64(mem.NumForcedGC)
	c.gaugeMetrics["NumGC"] = float64(mem.NumGC)
	c.gaugeMetrics["OtherSys"] = float64(mem.OtherSys)
	c.gaugeMetrics["PauseTotalNs"] = float64(mem.PauseTotalNs)
	c.gaugeMetrics["StackInuse"] = float64(mem.StackInuse)
	c.gaugeMetrics["StackSys"] = float64(mem.StackSys)
	c.gaugeMetrics["Sys"] = float64(mem.Sys)
	c.gaugeMetrics["TotalAlloc"] = float64(mem.TotalAlloc)
	c.gaugeMetrics["RandomValue"] = rand.Float64()

	for name, value := range c.gaugeMetrics {
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

func (c *Collector) getMetricsList() []domain.Metric {
	metricList := make([]domain.Metric, 0, len(c.metrics))
	for _, metric := range c.metrics {
		metricList = append(metricList, metric)
	}
	return metricList
}

func (c *Collector) GetMetrics() []domain.Metric {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getMetricsList()
}

func (c *Collector) Report() {
	c.mu.Lock()
	defer c.mu.Unlock()
	metricList := c.getMetricsList()
	c.report.MetricsReport(metricList)
}

type MetricsReporter interface {
	MetricsReport(metrics []domain.Metric)
}
