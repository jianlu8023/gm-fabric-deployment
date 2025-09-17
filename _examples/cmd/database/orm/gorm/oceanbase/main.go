package main

import (
	"fmt"
	"time"

	"_examples/internal/logger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "demo_app@demo:admin123$@tcp(192.168.58.110:2881)/demo?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.GetDBLogger(),
	})
	if err != nil {
		logger.GetAppLogger().Errorf(">>> get oceanbase db failed error :%v", err)
		return
	}

	if err = db.AutoMigrate(&User{}); err != nil {
		logger.GetAppLogger().Errorf(">>> auto migrate failed error :%v", err)
		return
	}

	defer func() {
		if r := recover(); r != nil {
			logger.GetAppLogger().Errorf(">>> recover error :%v", r)
		}
	}()
	defer func() {
		if err := db.Migrator().DropTable(&User{}); err != nil {
			logger.GetAppLogger().Errorf(">>> drop table failed error :%v", err)
			return
		}
	}()
	// 记录开始时间
	start := time.Now()

	user := User{Name: "OceanBase", Age: 12, Birthday: time.Date(2022, 06, 01, 00, 00, 00, 00, time.UTC)}
	result := db.Create(&user)
	fmt.Println(user.ID)
	fmt.Println(result.Error)
	fmt.Println(result.RowsAffected)

	user = User{ID: 1}
	result = db.First(&user)
	fmt.Println(user)
	fmt.Println(result.Error)
	fmt.Println(result.RowsAffected)

	user = User{ID: 1, Name: "ob", Age: 13, Birthday: time.Date(2023, 06, 01, 00, 00, 00, 00, time.UTC)}
	result = db.Save(&user)
	fmt.Println(user)
	fmt.Println(result.Error)
	fmt.Println(result.RowsAffected)

	user = User{ID: 1}
	result = db.Delete(&user)
	fmt.Println(user.ID)
	fmt.Println(result.Error)
	fmt.Println(result.RowsAffected)

	fmt.Println(time.Since(start))
}

type User struct {
	ID       int
	Name     string
	Age      int
	Birthday time.Time
}

func (User) TableName() string {
	return "t_user"
}
