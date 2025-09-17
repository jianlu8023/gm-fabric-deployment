package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/failsafe-go/failsafe-go"
	"github.com/failsafe-go/failsafe-go/failsafehttp"
)

// https://github.com/failsafe-go/failsafe-go

func main() {
	// Create a RetryPolicy that handles non-terminal responses
	retryPolicy := failsafehttp.RetryPolicyBuilder().
		OnRetryScheduled(func(e failsafe.ExecutionScheduledEvent[*http.Response]) {
			fmt.Println("Ping retry", e.Attempts(), "after delay of", e.Delay)
		}).Build()

	// Use the RetryPolicy with a failsafe RoundTripper

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
		}()
		fmt.Println("with failsafe round tripper")
		{

			// Setup a test http server that returns 429 on the first two requests with a 1 second Retry-After header
			server := flakyServer(2, 429, time.Second)
			defer server.Close()

			roundTripper := failsafehttp.NewRoundTripper(nil, retryPolicy)
			client := &http.Client{Transport: roundTripper}

			fmt.Println("Sending ping")
			req, _ := http.NewRequest(http.MethodGet, server.URL, strings.NewReader("ping"))
			resp, err := client.Do(req)

			readAndPrintResponse(resp, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		// Use the RetryPolicy with an HTTP client via a failsafe execution
		println("with failsafe execution")
		{
			// Setup a test http server that returns 429 on the first two requests with a 1 second Retry-After header
			server := flakyServer(2, 429, time.Second)
			defer server.Close()

			fmt.Println("Sending ping")
			resp, err := failsafe.GetWithExecution(func(exec failsafe.Execution[*http.Response]) (*http.Response, error) {
				// Include the execution context in the request, so that cancellations are propagated
				req, _ := http.NewRequestWithContext(exec.Context(), http.MethodGet, server.URL, strings.NewReader("ping"))
				client := &http.Client{}
				return client.Do(req)
			}, retryPolicy)

			readAndPrintResponse(resp, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()
		// Use the RetryPolicy with a failsafehttp.Request
		println("with failsafehttp.Request")
		{
			// Setup a test http server that returns 429 on the first two requests with a 1 second Retry-After header
			server := flakyServer(2, 429, time.Second)
			defer server.Close()
			client := &http.Client{}

			fmt.Println("Sending ping")
			req, _ := http.NewRequest(http.MethodGet, server.URL, strings.NewReader("ping"))
			failsafeReq := failsafehttp.NewRequest(req, client, retryPolicy)
			resp, err := failsafeReq.Do()

			readAndPrintResponse(resp, err)
		}

	}()
	wg.Wait()

}
func flakyServer(failTimes int, responseCode int, retryAfterDelay time.Duration) *httptest.Server {
	failures := atomic.Int32{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		body, _ := io.ReadAll(request.Body)
		fmt.Println("Received request", string(body))
		if failures.Add(1) <= int32(failTimes) {
			if retryAfterDelay > 0 {
				w.Header().Add("Retry-After", strconv.Itoa(int(retryAfterDelay.Seconds())))
			}
			fmt.Println("Replying with", responseCode)
			w.WriteHeader(responseCode)
		} else {
			fmt.Fprintf(w, "pong")
		}
	}))
}

func slowServer(delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		timer := time.After(delay)
		select {
		// request.Context() will be done as soon as the first successful response is handled by the client
		case <-request.Context().Done():
		case <-timer:
			fmt.Fprintf(w, "pong")
		}
	}))
}
func readAndPrintResponse(response *http.Response, err error) {
	if err != nil {
		fmt.Println("Received", err)
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	if len(body) > 0 {
		fmt.Println("Received", string(body))
	}
}
