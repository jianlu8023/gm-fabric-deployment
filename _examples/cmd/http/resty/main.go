package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"

	"_examples/internal/logger"
)

// https://github.com/go-resty/resty

type AuthSuccess struct {
	ID, Message string
}

type AuthError struct {
	ID, Message string
}

type User struct {
	Username string
	Password string
}

var (
	client = resty.New().
		SetTLSClientConfig(&tls.Config{
			InsecureSkipVerify: true,
		}).
		// Set retry-go count to non zero to enable retries
		SetRetryCount(3).
		// You can override initial retry-go wait time.
		// Default is 100 milliseconds.
		SetRetryWaitTime(5 * time.Second).
		// MaxWaitTime can be overridden as well.
		// Default is 2 seconds.
		SetRetryMaxWaitTime(20 * time.Second).
		// SetRetryAfter sets callback to calculate wait time between retries.
		// Default (nil) implies exponential backoff with jitter
		SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
			return 0, errors.New("quota exceeded")
		})
)

func main() {

	resp, err := client.R().
		EnableTrace().
		Get("https://httpbin.org/get")

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()

	// Explore trace info
	fmt.Println("Request Trace Info:")
	ti := resp.Request.TraceInfo()
	fmt.Println("  DNSLookup     :", ti.DNSLookup)
	fmt.Println("  ConnTime      :", ti.ConnTime)
	fmt.Println("  TCPConnTime   :", ti.TCPConnTime)
	fmt.Println("  TLSHandshake  :", ti.TLSHandshake)
	fmt.Println("  ServerTime    :", ti.ServerTime)
	fmt.Println("  ResponseTime  :", ti.ResponseTime)
	fmt.Println("  TotalTime     :", ti.TotalTime)
	fmt.Println("  IsConnReused  :", ti.IsConnReused)
	fmt.Println("  IsConnWasIdle :", ti.IsConnWasIdle)
	fmt.Println("  ConnIdleTime  :", ti.ConnIdleTime)
	fmt.Println("  RequestAttempt:", ti.RequestAttempt)
	fmt.Println("  RemoteAddr    :", ti.RemoteAddr.String())
	// POST JSON string
	// No need to set content type, if you have client level setting

	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"username":"testuser", "password":"testpass"}`).
		SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		Post("https://myapp.com/login")

	// POST []byte array
	// No need to set content type, if you have client level setting
	resp, err = client.R().
		SetHeader("Content-Type", "application/json").
		SetBody([]byte(`{"username":"testuser", "password":"testpass"}`)).
		SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		Post("https://myapp.com/login")

	// POST Struct, default is JSON content type. No need to set one
	resp, err = client.R().
		SetBody(User{Username: "testuser", Password: "testpass"}).
		SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		SetError(&AuthError{}).    // or SetError(AuthError{}).
		Post("https://myapp.com/login")

	// POST Map, default is JSON content type. No need to set one
	resp, err = client.R().
		SetBody(map[string]interface{}{"username": "testuser", "password": "testpass"}).
		SetResult(&AuthSuccess{}). // or SetResult(AuthSuccess{}).
		SetError(&AuthError{}).    // or SetError(AuthError{}).
		Post("https://myapp.com/login")

	resp, err = client.R().Post("https://172.25.138.18:7803/exchange")
	if err != nil {
		logger.GetAppLogger().Errorf("发送post请求失败：%v", err)
		return
	}

	logger.GetAppLogger().Debugf("resp :%v", string(resp.Body()))
	client.AddRetryCondition(
		// RetryConditionFunc type is for retry-go condition function
		// input: non-nil Response OR request execution error
		func(r *resty.Response, err error) bool {
			return r.StatusCode() == http.StatusTooManyRequests
		},
	)
}
