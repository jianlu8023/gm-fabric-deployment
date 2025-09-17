package main

import (
	"_examples/internal/database/entity"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"_examples/internal/logger"
)

const (
	url = "%v:%v@tcp(%v:%v)/%v?charset=utf8mb4&parseTime=True&loc=Local"
)

func main() {

	sprintf := fmt.Sprintf(url, "root", "123456", "127.0.0.1", "3306", "basic")

	db, err := gorm.Open(mysql.Open(sprintf), &gorm.Config{
		PrepareStmt: true,
		Logger:      logger.GetDBLogger(),
	})
	if err != nil {
		log.Fatalln(err)
		return
	}
	if err = db.AutoMigrate(&entity.User{}); err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 10; i++ {
		if err = db.Model(&entity.User{}).Create(&entity.User{
			UserName: fmt.Sprintf("user_%v", i),
		}).Error; err != nil {
			fmt.Println(err)
			return
		}
	}

	var version string
	db.Raw("SELECT VERSION()").Row().Scan(&version)
	fmt.Println("数据库版本:", version)
}
