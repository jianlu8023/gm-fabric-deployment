package main

import (
	"context"
	"fmt"
	"log"

	"github.com/brianvoe/gofakeit/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// https://www.mongodb.com/zh-cn/docs/

func main() {
	// 设置 MongoDB 连接选项
	clientOptions := options.Client().ApplyURI("mongodb://root:123456@localhost:27017")

	// 连接到 MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// 检查连接
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	// 获取数据库和集合
	collection := client.Database("basic").Collection("user")

	// 插入一条文档
	doc := bson.D{{"name", gofakeit.FirstName()}, {"age", gofakeit.Number(0, 90)}}
	insertResult, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	// 查询所有文档
	cursor, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var result bson.D
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}

	// 断开与 MongoDB 的连接
	err = client.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
}
