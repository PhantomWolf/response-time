package main

import (
	"fmt"
	flag "github.com/spf13/pflag"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type result struct {
	Seconds float64
	Err     error
}

// checkOnce sends a single request, reads the response body, and returns the time used
func checkOnce(client *http.Client, req *http.Request) (float64, error) {
	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("%s %s: %s\n", req.Method, req.URL.String(), err.Error())
		return -1, err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body) // Read response body
	secs := time.Since(start).Seconds()
	if err != nil {
		log.Printf("Reading response body failed: %s\n", err.Error())
		return secs, err
	}
	log.Printf("%s %s: %s(%.2f secs)\n", req.Method, req.URL.String(), resp.Status, secs)
	return secs, nil
}

func usage() string {
	return fmt.Sprintf("Usage: %s [--method METHOD] [--data DATA] URL", os.Args[0])
}

func main() {
	// Parse command-line arguments
	var config struct {
		Method   string
		Data     string
		URL      string
		Interval int64
		Time     int64
	}
	flag.StringVarP(&config.Method, "request", "X", "GET", "Use http method `METHOD`")
	flag.StringVarP(&config.Data, "data", "d", "", "Send data `DATA` in http message body")
	flag.Int64VarP(&config.Interval, "interval", "i", 10, "Check url status every `INTERVAL` seconds")
	flag.Int64VarP(&config.Time, "time", "t", 5, "Run for `MINS` minutes in total")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalln(usage())
	}
	config.URL = flag.Arg(0)

	// Create HTTP request
	var body io.Reader
	if len(config.Data) != 0 {
		body = strings.NewReader(config.Data)
	}
	req, err := http.NewRequest(config.Method, config.URL, body)
	if err != nil {
		log.Fatalf("Invalid request: %s\n", err.Error())
	}
	req.Header.Set("Cache-Control", "no-store") // Disable cache

	// Create HTTP client
	transport := &http.Transport{}
	client := &http.Client{Transport: transport}

	// Send requests in another goroutine
	ch := make(chan *result, 10)
	ticker := time.NewTicker(time.Second * time.Duration(config.Interval))
	timer := time.NewTimer(time.Minute * time.Duration(config.Time))
	go func() {
		for {
			select {
			case <-timer.C:
				return
			case <-ticker.C:
				secs, err := checkOnce(client, req)
				ch <- &result{Seconds: secs, Err: err}
			}
		}
	}()

	// Statistics
	totalReq := 0
	var averageTime float64 = 0
	var totalTime float64 = 0
	for {
		select {
		case <-timer.C:
			averageTime = totalTime / float64(totalReq)
		case res := <-ch:
			totalTime += res.Seconds
			totalReq += 1
			continue
		}
		break
	}
	fmt.Printf("Average response time: %f\n", averageTime)
}
