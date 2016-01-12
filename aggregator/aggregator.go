package main

import (
	"bufio"
	"os"
	"net/http"
	"time"
	"fmt"
	"errors"
	"encoding/json"
)


type BreakerCount struct {
	OpenCount   int64
	ClosedCount int64
}

type LatencyHistogram struct {
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

type CircuitBreaker struct {
	Name              string
	SuccessCount      int64
	FailCount         int64
	ShortCircuitCount int64
	WindowDuration    time.Duration
	CurrentTime       time.Time
	BreakerStatus     BreakerCount
	Latency           LatencyHistogram
}

type SSEString string

// A snapshot transcription of the hystrix.stream JSON object
// This is here for legacy support only. Only update if the fields change or
// In the event of an inevitable bug.
type HystrixStream struct {
	// Forgive my ridiculous formatting in this ridiculous object
	CurrentConcurrentExecutionCount int
	CurrentTime          string            `json:"currentTime,string"`
	ErrorPercentage      int               `json:"errorPercentage,int"`
	ErrorCount           int               `json:"errorCount,int"`
	Group                string            `json:"group,string"`
	IsCircuitBreakerOpen bool              `json:"isCircuitBreakerOpen,bool"`
	LatencyExecute       HystrixHistogram  `json:"latencyExecute,HystrixHistogram"`
	LatencyExecuteMean   int               `json:"latencyExecute_mean,int"`
	LatencyTotal         HystrixHistogram  `json:"latencyTotal,HystrixHistogram"`
	LatencyTotalMean     int               `json:"latencyTotal_mean,int"`
	Name                 string            `json:"name,string"`
	ReportingHosts       int               `json:"reportingHosts,int"`
	RequestCount         int               `json:"requestCount,int"`
	RollingCountCollapsedRequests   int    `json:"rollingCountCollapsedRequests,int"`
	RollingCountExceptionsThrown    int    `json:"rollingCountExceptionsThrown,int"`
	RollingCountFailure             int    `json:"rollingCountFailure,int"`
	RollingCountFallbackFailure     int    `json:"rollingCountFallbackFailure,int"`
	RollingCountFallbackRejection   int    `json:"rollingCountFallbackRejection,int"`
	RollingCountResponseFromCache   int    `json:"rollingCountResponseFromCache,int"`
	RollingCountSemaphoreRejected   int    `json:"rollingCountSemaphoreRejected,int"`
	RollingCountShortCircuited      int    `json:"rollingCountShortCircuited,int"`
	RollingCountSuccess             int    `json:"rollingCountSuccess,int"`
	RollingCountThreadPoolRejected  int    `json:"rollingCountThreadPoolRejected,int"`
	RollingCountTimeout             int    `json:"rollingCOuntTimeout,int"`
	Type                            string `json:"type,string"`
	// Don't blame me for these awful names.
	// I'm preserving the bad names Hystrix uses
	PropertyValueCircuitBreakerEnabled                            bool   `json:"propertyValue_circuitBreakerEnabled,bool"`
	PropertyValueCircuitBreakerErrorThresholdPercentage           int    `json:"propertyValue_circuitBreakerErrorThresholdPercentage,int"`
	PropertyValueCircuitBreakerForceOpen                          bool   `json:"propertyValue_circuitBreakerForceOpen,bool"`
	PropertyValueCircuitBreakerForceClosed                        bool   `json:"propertyValue_circuitBreakerForceClosed,bool"`
	PropertyValueCircuitBreakerRequestVolumeThreshold             int    `json:"propertyValue_circuitBreakerRequestVolumeThreshold,int"`
	PropertyValueCircuitBreakerSleepWindowInMilliseconds          int    `json:"propertyValue_circuitBreakerSleepWindowInMilliseconds,int"`
	PropertyValueExecutionIsolationSemaphoreMaxConcurrentRequests int    `json:"propertyValue_executionIsolationSemaphoreMaxConcurrentRequests,int"`
	PropertyValueExecutionIsolationStrategy                       string `json:"propertyValue_executionIsolationStrategy,string"`
	PropertyValueExecutionIsolationThreadPoolKeyOverride          string `json:"propertyValue_executionIsolationThreadPoolKeyOverride,string"`
	PropertyValueExecutionIsolationThreadTimeoutInMilliseconds    int    `json:"propertyValue_executionIsolationThreadTimeoutInMilliseconds,string"`
	PropertyValueFallbackIsolationSemaphoreMaxConcurrentRequests  int    `json:"propertyValue_fallbackIsolationSeampahoreMaxConcurrentRequests,int"`
	PropertyValueMetricsRollingStatisticalWindowInMilliseconds    int    `json:"propertyValue_metricsRollingStatisticalWindowInMilliseconds,int"`
	PropertyValueRequestCacheEnabled                              bool   `json:"propertyValue_requestCacheEnabled,bool"`
	PropertyValueRequestLogEnabled                                bool   `json:"propertyValue_requestLogEnabled,bool"`
}

// A snapshot transcription of the histogram objects hystrix.stream JSON object
// This is here for legacy support only. Only update if the fields change or
// In the event of an inevitable bug.
type HystrixHistogram struct {
	//minimum
	percentile0   int `json:"0,string"`
	percentile25  int `json:"25,string"`
	//median
	percentile50  int `json:"50,string"`
	percentile75  int `json:"75,string"`
	percentile90  int `json:"90,string"`
	percentile95  int `json:"95,string"`
	percentile995 int `json:"99.5,string"`
	//maximum
	percentile100 int `json:"100,string"`
}

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

func (h HystrixStream) ToCircuitBreaker() (CircuitBreaker, error) {
	return CircuitBreaker{}, nil
}

func (c CircuitBreaker) ToJSON() string {
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
		c.Name,
		c.SuccessCount,
		c.FailCount,
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
	stdout :=  bufio.NewWriter(os.Stdout)
	stdout.ReadFrom(resp.Body)
	defer stdout.Flush()
}


