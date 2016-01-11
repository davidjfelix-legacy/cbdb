package main

import (
	"bufio"
	"os"
	"net/http"
	"time"
	"fmt"
)


type breakerCount struct {
	openCount   int64
	closedCount int64
}

type latencyHistogram struct {
	mean          int64
	median        int64
	min           int64
	max           int64
	percentile25  int64
	percentile75  int64
	percentile90  int64
	percentile95  int64
	percentile99  int64
	percentile995 int64
	percentile999 int64
}

type CircuitBreaker struct{
	name              string
	successCount      int64
	failCount         int64
	shortCircuitCount int64
	windowDuration    time.Duration
	currentTime       time.Time
	breakerStatus     breakerCount
	latency           latencyHistogram
}

type SSEString string

func (s *SSEString) CircuitBreakerFromHystrix() CircuitBreaker, error {
}

func (c *CircuitBreaker) ToJSON() string {
	return fmt.Sprintf(
		"{" +
			"\"name\":%v," +
			"\"successCount\":%v," +
			"\"failCount\":%v," +
			"\"shortCircuitCount\":%v," +
			"\"windowDuration\":%v," +
			"\"currentTime\":%v," +
			"\"breakerStatus\":%v," +
			"\"latency\":%v" +
		"}",
		c.name,
		c.successCount,
		c.failCount,
		c.shortCircuitCount,
		(c.windowDuration.Nanoseconds() / 1000),
		c.currentTime.Format(time.RFC3339),
		c.breakerStatus.toJSON(),
		c.latency.toJSON())
}

func (l *latencyHistogram) toJSON() string {
	return fmt.Sprintf(
		"{" +
			"\"mean\":%v," +
			"\"median\":%v," +
			"\"min\":%v," +
			"\"max\":%v,"+
			"\"25\":%v," +
			"\"75\":%v," +
			"\"90\":%v," +
			"\"95\":%v," +
			"\"99\":%v," +
			"\"99.5\"%v," +
			"\"99.9\"%v" +
		"}",
		l.mean,
		l.median,
		l.min,
		l.max,
		l.percentile25,
		l.percentile75,
		l.percentile90,
		l.percentile95,
		l.percentile99,
		l.percentile995,
		l.percentile999)
}

func (b *breakerCount) toJSON() string {
	return fmt.Sprintf(
		"{" +
			"\"open\":%v," +
			"\"closed\":%v" +
		"}",
		b.openCount,
		b.closedCount)
}

func main() {
	req, err := http.NewRequest("GET", "/hystrix.stream", nil)
	if err != nil {
		os.Exit(1)
	}
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		os.Exit(1)
	}
	stdout :=  bufio.NewWriter(os.Stdout)
	stdout.ReadFrom(resp.Body)
	defer stdout.Flush()
}


