package main

import (
	"sync"
	"bufio"
	"os"
	"net/http"
	"time"
	"fmt"
	"errors"
	"encoding/json"
	"strconv"
)


type BreakerCount struct {
	OpenCount   int64
	ClosedCount int64
}

type LatencyHistogram struct {
	Mean          int64
	Median        int64
	Min           int64
	Max           int64
	Percentile25  int64
	Percentile75  int64
	Percentile90  int64
	Percentile95  int64
	Percentile99  int64
	Percentile995 int64
	Percentile999 int64
}

type CircuitBreaker struct {
	Name              string
	SuccessCount      int64
	FailCount         int64
	FallbackCount     int64
	ShortCircuitCount int64
	WindowDuration    time.Duration
	CurrentTime       time.Time
	BreakerStatus     BreakerCount
	Latency           LatencyHistogram
}



func (c CircuitBreaker) ToJSON() string {
	return fmt.Sprintf(
		"{" +
			"\"name\":%v," +
			"\"successCount\":%v," +
			"\"failCount\":%v," +
			"\"fallbackCount\":%v," +
			"\"shortCircuitCount\":%v," +
			"\"windowDuration\":%v," +
			"\"currentTime\":%v," +
			"\"breakerStatus\":%v," +
			"\"latency\":%v" +
		"}",
		c.Name,
		c.SuccessCount,
		c.FailCount,
		c.FallbackCount,
		c.ShortCircuitCount,
		(c.WindowDuration.Nanoseconds() / 1000),
		c.CurrentTime.Format(time.RFC3339),
		c.BreakerStatus.toJSON(),
		c.Latency.toJSON())
}

func (l LatencyHistogram) toJSON() string {
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
		l.Mean,
		l.Median,
		l.Min,
		l.Max,
		l.Percentile25,
		l.Percentile75,
		l.Percentile90,
		l.Percentile95,
		l.Percentile99,
		l.Percentile995,
		l.Percentile999)
}

func (b BreakerCount) toJSON() string {
	return fmt.Sprintf(
		"{" +
			"\"open\":%v," +
			"\"closed\":%v" +
		"}",
		b.OpenCount,
		b.ClosedCount)
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
	var wg sync.WaitGroup
	c := make(chan rune)
	stdout := bufio.NewWriter(os.Stdout)
	bodyin := bufio.NewReader(resp.Body)
	wg.Add(2)
	//FIXME: need to handle cleanup of the channel and waitgroup
	go func() {
		//FIXME: This doesn't check errors at all or break cleanly
		defer wg.Done()
		for {
			r, _, _ := bodyin.ReadRune()
			c <- r
		}
	}()
	go func() {
		//FIXME: This is probably not safe
		defer wg.Done()
		for {
			r := <-c
			stdout.WriteRune(r)
			stdout.Flush()
		}
	}()
	wg.Wait()
}


