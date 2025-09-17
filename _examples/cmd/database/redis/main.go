package main

import (
	"context"
	"time"

	gredis "github.com/redis/go-redis/v9"

	"_examples/internal/logger"
)

var (
	log = logger.GetAppLogger()
)

func main() {

	var (
		ctx = context.Background()
	)

	rds := gredis.NewClient(&gredis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
		DB:       5,
	})

	if ping, err := rds.Ping(ctx).Result(); err != nil {
		log.Errorf("ping error %v", err)
	} else {
		log.Infof("ping result %v", ping)
	}

	time.Sleep(5 * time.Second)
	key := "testKey"
	value := "testValue"
	nvalue := "testValue-new"
	if set, err := rds.Set(ctx, key, value, 0).Result(); err != nil {
		log.Errorf("set error %v", err)
	} else {
		log.Infof("set result %v", set)
	}
	time.Sleep(5 * time.Second)

	if dbsize, err := rds.DBSize(ctx).Result(); err != nil {
		log.Errorf("dbsize error %v", err)
	} else {
		log.Infof("dbsize result %v", dbsize)
	}

	time.Sleep(5 * time.Second)
	if get, err := rds.Get(ctx, key).Result(); err != nil {
		log.Errorf("get error %v", err)
	} else {
		log.Infof("get result %v", get)
	}
	time.Sleep(5 * time.Second)

	if rset, err := rds.Set(ctx, key, nvalue, 0).Result(); err != nil {
		log.Errorf("rset error %v", err)
	} else {
		log.Infof("rset result %v", rset)
	}
	time.Sleep(5 * time.Second)

	if rget, err := rds.Get(ctx, key).Result(); err != nil {
		log.Errorf("rget error %v", err)
	} else {
		log.Infof("rget result %v", rget)
	}
	time.Sleep(5 * time.Second)

	if del, err := rds.Del(ctx, key).Result(); err != nil {
		log.Errorf("del error %v", err)
	} else {
		log.Infof("del result %v", del)
	}
}
