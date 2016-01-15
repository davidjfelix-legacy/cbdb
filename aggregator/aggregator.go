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

type SSEString string


func (s SSEString) ParseHystrixStream() (HystrixStream, error) {
	// The eventsource string isn't big short circuit
	if len(s) < 8 {
		return HystrixStream{}, errors.New("Event string too short to parse")
	}

	// The eventsource string isn't data
	if s[:6] != "data: " {
		return HystrixStream{}, errors.New("Can't parse non-data event")
	}

	// Try to parse JSON
	var ret HystrixStream
	resp := json.Unmarshal([]byte(s[7:]), &ret)

	if resp == nil {
		return ret, nil
	} else {
		return HystrixStream{}, resp
	}
}

func (h HystrixHistogram) ToLatencyHistogram(mean int64) LatencyHistogram {
	return LatencyHistogram {
		Mean: mean,
		Median: h.Percentile50,
		Min: h.Percentile0,
		Max: h.Percentile100,
		Percentile25: h.Percentile25,
		Percentile75: h.Percentile75,
		Percentile90: h.Percentile90,
		Percentile95: h.Percentile95,
		Percentile99: h.Percentile99,
		Percentile995: h.Percentile995,
		// Unfortunately, the closest we have is an estimate between 99.5 and 100. We'll take it
		Percentile999: (h.Percentile100 + h.Percentile995) / 2,
	}
}

func (h HystrixStream) ToCircuitBreaker() (CircuitBreaker, error) {
	var breakerCount BreakerCount
	if h.IsCircuitBreakerOpen {
		breakerCount = BreakerCount{OpenCount: 1, ClosedCount: 0}
	} else {
		breakerCount = BreakerCount{OpenCount: 0, ClosedCount: 1}
	}

	var currentTime time.Time

	// This is how I parse the time.
	parsedTime, err := strconv.Atoi(h.CurrentTime)
	if err != nil {
		return CircuitBreaker{}, err
	} else {
		// Split hystrix ms encoded unix time into s and ns
		currentTime = time.Unix(int64(parsedTime / 1000), int64((parsedTime % 1000) * 1000))
	}

	return CircuitBreaker {
		Name: h.Group + h.Name,
		SuccessCount: h.RollingCountSuccess,
		FailCount: 1,
		FallbackCount: 1,
		ShortCircuitCount: 1,
		WindowDuration: 1,
		CurrentTime: currentTime,
		BreakerStatus: breakerCount,
		Latency: h.LatencyTotal.ToLatencyHistogram(h.LatencyTotalMean),
	}, nil
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


