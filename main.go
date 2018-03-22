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

type Config struct {
	Method   string
	Data     string
	URL      string
	Interval int64
	Time     int64
	Follow   bool
}

type Result struct {
	Seconds float64
	Err     error
}

// CheckOnce sends a single request, reads the response body, and returns the time used
func CheckOnce(client *http.Client, req *http.Request) (float64, error) {
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

func Check(client *http.Client, req *http.Request, timer *time.Timer, ticker *time.Ticker, ch chan<- *Result) {
	for {
		select {
		case <-timer.C:
			return
		case <-ticker.C:
			secs, err := CheckOnce(client, req)
			ch <- &Result{Seconds: secs, Err: err}
		}
	}
}

func Statistics(timer *time.Timer, ch <-chan *Result) float64 {
	count := 0
	var avg float64 = 0
	var total float64 = 0
	for {
		select {
		case <-timer.C:
			avg = total / float64(count)
			return avg
		case res := <-ch:
			total += res.Seconds
			count += 1
		}
	}
}

func Usage() string {
	return fmt.Sprintf("Usage: %s [--method METHOD] [--data DATA] URL", os.Args[0])
}

func parseArgs() *Config {
	config := &Config{}
	flag.StringVarP(&config.Method, "request", "X", "GET", "Use http method `METHOD`")
	flag.StringVarP(&config.Data, "data", "d", "", "Send data `DATA` in http message body")
	flag.Int64VarP(&config.Interval, "interval", "i", 10, "Check url status every `INTERVAL` seconds")
	flag.Int64VarP(&config.Time, "time", "t", 5, "Run for `MINS` minutes in total")
	flag.BoolVarP(&config.Follow, "location", "L", false, "Follow redirects")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatalln(Usage())
	}
	config.URL = flag.Arg(0)
	if !strings.HasPrefix(config.URL, "http") {
		config.URL = "http://" + config.URL
	}
	return config
}

func NewHTTPRequest(method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Cache-Control", "no-store") // Disable cache
	return req, nil
}

func NewHTTPClient(follow bool) *http.Client {
	// Do not follow redirections
	client := &http.Client{}
	if !follow {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client
}

func main() {
	config := parseArgs()

	var body io.Reader
	if len(config.Data) != 0 {
		body = strings.NewReader(config.Data)
	}
	req, err := NewHTTPRequest(config.Method, config.URL, body)
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %s\n", err.Error())
	}

	// Send requests in another goroutine
	client := NewHTTPClient(config.Follow)
	ch := make(chan *Result, 10)
	ticker := time.NewTicker(time.Second * time.Duration(config.Interval))
	timer := time.NewTimer(time.Minute * time.Duration(config.Time))
	go Check(client, req, timer, ticker, ch)
	// Statistics
	avg := Statistics(timer, ch)
	fmt.Printf("Average response time: %.2fs\n", avg)
}
