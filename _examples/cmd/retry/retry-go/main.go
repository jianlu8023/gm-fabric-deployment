package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/avast/retry-go/v4"
	// "github.com/stretchr/testify/assert"
)

func main() {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello")
	}))
	defer ts.Close()

	var body []byte

	err := retry.Do(
		func() error {
			resp, err := http.Get(ts.URL)

			if err == nil {
				defer func() {
					if err := resp.Body.Close(); err != nil {
						panic(err)
					}
				}()
				body, err = io.ReadAll(resp.Body)
			}

			return err
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(body))
	// assert.NoError(t, err)
	// assert.NotEmpty(t, body)
}
