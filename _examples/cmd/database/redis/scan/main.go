package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	gredis "github.com/redis/go-redis/v9"

	"_examples/internal/logger"
)

var (
	g         singleflight.Group
	wg        sync.WaitGroup
	semaphore = make(chan struct{}, 20)
	ctx       = context.Background()
	log       = logger.GetAppLogger()
)

func main() {

	rds := gredis.NewClient(&gredis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       1,
	})

	var (
		err    error
		keys   []string
		cursor uint64
	)

	for {
		keys, cursor, err = rds.Scan(ctx, cursor, "*", 0).Result()
		if err != nil {
			log.Errorf("从redis执行scan失败 %v", err)
			return
		}

		for _, key := range keys {
			semaphore <- struct{}{}
			wg.Add(1)
			goroutineID := fmt.Sprintf("R-%v", key)
			log.Infof("获取到信号量启动协程 %v ", goroutineID)

			go func(k, goroutineID string) {
				defer func() {
					<-semaphore
					wg.Done()
					log.Debugf("协程 %v 执行完任务,释放信号量...", goroutineID)
				}()

				defer func() {
					if err := recover(); err != nil {
						log.Errorf("协程执行过程中发生错误 %v", err)
					}
				}()

				time.Sleep(time.Duration(rand.Intn(80)) * time.Second)

				do, err, shared := g.Do(k, func() (interface{}, error) {
					result, err := rds.Get(ctx, k).Result()
					if err != nil {
						return "", err
					} else {
						return result, nil
					}
				})
				if err != nil {
					log.Errorf("g.Do 发生错误 %v", err)
				}
				if shared {
					log.Debugf("共享的数据 %v", do.(string))
				} else {
					log.Infof("非共享的数据 %v", do.(string))
				}

			}(key, goroutineID)

		}
		if cursor == 0 {
			break
		}
	}
	wg.Wait()
}
