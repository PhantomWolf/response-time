package main

import (
	"log"
	"math"
	"net/http"
	"testing"
	"time"
)

const (
	responseTime = 1.0
)

func init() {
	server := &http.Server{Addr: "127.0.0.1:8080"}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * responseTime)
		s := "good"
		w.Write([]byte(s))
	})
	go func() {
		switch err := server.ListenAndServe(); err {
		case http.ErrServerClosed:
			log.Println("HTTP server closed")
		default:
			log.Fatalf("Failed to start mock HTTP server: %s\n", err.Error())
		}
	}()
}

func TestCheckOnce(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
	secs, err := CheckOnce(client, req)
	if err != nil {
		t.Fatalf("Check url failed: %s\n", err.Error())
	}
	// Response time should be roughly 3 seconds
	if math.Abs(secs-responseTime) > 0.2 {
		t.Fatalf("Response time incorrect. Expected: %.2f, Actual: %.2f\n", responseTime, secs)
	}
	t.Logf("Response time: %.2f\n", secs)
}

func TestCheck(t *testing.T) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
	timer := time.NewTimer(time.Second * 5)
	ticker := time.NewTicker(time.Second)
	ch := make(chan *Result, 10)
	go Check(client, req, timer, ticker, ch)
	for {
		select {
		case <-timer.C:
		case res := <-ch:
			if res.Err != nil {
				t.Fatalf("GET http://127.0.0.1:8080 failed: %s\n", res.Err.Error())
			}
			if math.Abs(res.Seconds-responseTime) > 0.2 {
				t.Fatalf("Response time incorrect. Expected: %.2f, Actual: %.2f\n", responseTime, res.Seconds)
			}
			continue
		}
		break
	}
	t.Logf("Test passed")
}

func TestStatistics(t *testing.T) {
	timer := time.NewTimer(time.Second * 5)
	ch := make(chan *Result, 10)

	go func() {
		for i := 1; i <= 10; i++ {
			result := &Result{Seconds: float64(i), Err: nil}
			ch <- result
		}
	}()

	avg := Statistics(timer, ch)
	if avg != 5.5 {
		t.Fatalf("Average response time incorrect. Expected: 5.5, Actual: %.2f\n", avg)
	}
}
