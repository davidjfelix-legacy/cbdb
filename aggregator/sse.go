package main

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
