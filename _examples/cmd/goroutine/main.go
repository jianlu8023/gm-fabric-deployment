package main

import (
	"math/rand/v2"
	"sync"
	"time"

	"_examples/internal/logger"
)

var (
	log = logger.GetAppLogger()
)

func main() {

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 20)
	go func() {

		NewTicker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-NewTicker.C:
				l := len(semaphore)
				c := cap(semaphore)
				log.Infof(">>> daemon length %v cap %v", l, c)
			default:

			}
		}
	}()

	for i := 0; i <= 5; i++ {
		log.Infof(">>> 开始第 %v 轮数据...", i)
		for k := 1; k <= 30; k++ {
			semaphore <- struct{}{}
			wg.Add(1)
			log.Infof(">>> 第 %v 循环 协程 %v 获取信号量成功,开始执行任务...", i, k)
			go func(c, index int) {
				defer func() {
					<-semaphore
					wg.Done()
					log.Infof(">>> 第 %v 循环 协程 %v 完成任务 释放信号量...", c, index)
				}()
				sleep := time.Duration(rand.IntN(100)) * time.Second
				log.Infof(">>> 第 %v 循环 协程 %v start sleep %v", c, index, sleep)
				time.Sleep(sleep)
				log.Infof(">>> 第 %v 循环 协程 %v sleep %v end ", c, index, sleep)
			}(i, k)
		}

	}
	wg.Wait()
	log.Infof(">>> test finish ...")
}
