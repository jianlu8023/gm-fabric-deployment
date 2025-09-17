package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/ipfs/boxo/gateway"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"

	"_examples/internal/logger"
)

var appLogger = logger.GetAppLogger()

func main() {

	client := http.Client{
		// Timeout: 5 * time.Second,
	}
	backend, err := gateway.NewRemoteCarBackend([]string{"http://localhost:8080"}, &client)
	if err != nil {
		appLogger.Errorf("NewRemoteCarBackend Error: %v", err)
		return
	}
	ctx := context.Background()
	m := map[string]string{
		"bafybeigggyffcf6yfhx5irtwzx3cgnk6n3dwylkvcpckzhqqrigsxowjwe": "0.28.0webui.car",
		"bafybeif4zkmu7qdhkpf3pnhwxipylqleof7rl6ojbe7mq3fzogz6m4xk3i": "0.8.0webui.car",
		"bafybeihatzsgposbr3hrngo42yckdyqcc56yean2rynnwpzxstvdlphxf4": "0.30.0webui.car",
	}
	for k, v := range m {
		c, err := cid.Parse(k)
		if err != nil {
			appLogger.Errorf("parse cid Error %v", err)
			return
		}
		immutablePath, err := path.NewImmutablePath(path.FromCid(c))
		if err != nil {
			appLogger.Errorf("p.NewImmutablePath Error: %v", err)
			return
		}
		car, closer, err := backend.GetCAR(ctx, immutablePath, gateway.CarParams{
			Scope: gateway.DagScopeAll,
		})
		if err != nil {
			appLogger.Errorf("GetCAR Error: %v", err)
			return
		}
		appLogger.Infof("car: %v, closer: %v", car, closer)
		// defer func(closer io.ReadCloser) {
		// 	err := closer.Close()
		// 	if err != nil {
		// 		appLogger.Errorf("closer.Close Error: %v", err)
		// 	}
		// }(closer)

		p := "./testdata/" + v
		dir := filepath.Dir(p)
		err = os.MkdirAll(dir, os.ModeDir)
		if err != nil {
			appLogger.Errorf("os.MkdirAll Error: %v", err)
			return
		}

		create, err := os.Create(p)
		if err != nil {
			appLogger.Errorf("os.Create Error: %v", err)
			return
		}
		_, err = io.Copy(create, closer)
		if err != nil {
			appLogger.Errorf("io.Copy Error: %v", err)
			return
		}
		err = closer.Close()
		if err != nil {
			appLogger.Errorf("closer.Close Error: %v", err)
		}

		appLogger.Infof("get %v car success.", v)
	}

	appLogger.Infof("all car get it.")
}
