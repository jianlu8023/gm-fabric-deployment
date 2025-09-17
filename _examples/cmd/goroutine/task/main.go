package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"

	"_examples/internal/logger"
)

var log = logger.GetAppLogger()
var dataList [1000]int
var wg sync.WaitGroup
var semaphore = make(chan struct{}, 10)

func processData(d int) {
	defer func() {
		<-semaphore
		wg.Done()
	}()
	// time.Sleep(3 * time.Second)
	// 这里是处理数据d的具体逻辑，比如计算、存储等操作
	log.Debugf(">>> %d", d)
}
func processBatchData(dataBatch []int) {
	defer func() {
		<-semaphore
		wg.Done()
	}()
	for _, d := range dataBatch {
		// time.Sleep(3 * time.Second)
		// 这里是处理数据d的具体逻辑，比如计算、存储等操作
		log.Debugf(">>> %d", d)
	}
}

func main() {
	go func() {
		ticker := time.NewTicker(time.Second)

		for range ticker.C {
			log.Infof(">>> goroutine count: %d", runtime.NumGoroutine())
		}
	}()
	begin := time.Now()
	for _, d := range dataList {
		semaphore <- struct{}{}
		wg.Add(1)
		go processData(d)
	}
	wg.Wait()
	singleUse := time.Now()
	batchSize := 100
	for i := 0; i < len(dataList); i += batchSize {
		endIndex := i + batchSize
		if endIndex > len(dataList) {
			endIndex = len(dataList)
		}
		semaphore <- struct{}{}
		wg.Add(1)
		go processBatchData(dataList[i:endIndex])
	}
	wg.Wait()
	multiUse := time.Now()
	fmt.Println("one goroutine one data use time:", singleUse.Sub(begin))
	fmt.Println("one goroutine multi data use time:", multiUse.Sub(singleUse))
}
