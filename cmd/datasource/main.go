package main

import (
	"context"
	"fmt"

	"github.com/jianlu8023/go-tools/v2/pkg/helper/json"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/control/flags"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/version"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func (p Product) TableName() string {
	return "t_product"
}

func (p Product) String() string {
	bytes, _ := json.Marshal(p)
	return string(bytes)
}

func main() {

	flagsControl := flags.NewFlagsControl(version.Version)
	flagsControl.StartUp(func(err error) {
		fmt.Printf("start up flags control failed: %v\n", err)
		return
	})

	defer flagsControl.Shutdown()

	configControl := config.NewConfigControl(flagsControl)
	configControl.StartUp(func(err error) {
		fmt.Printf("start up config control failed: %v\n", err)
		return
	})

	defer configControl.Shutdown()

	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	loggerControl.StartUp(func(err error) {
		fmt.Printf("start up logger control failed: %v\n", err)
		return
	})
	defer loggerControl.Shutdown()

	datasourceConfig := &config.DataSourceConfig{
		Host:           "localhost",
		Port:           3306,
		DataSourceType: "sqlite3",
		UserName:       "root",
		Password:       "123456",
		DataBaseName:   "basic",
		DataBasePath:   "./db/golang-example.db",
		MaxIdleConn:    10,
		MaxOpenConn:    50,
	}

	dataSourceControl, err := datasource.NewDataSourceControl(datasourceConfig, loggerControl)
	if err != nil {
		fmt.Printf("init datasource failed: %v\n", err)
		return
	}
	dataSourceControl.StartUp(func(err error) {
		fmt.Printf("start up datasource failed: %v\n", err)
		return
	})
	defer dataSourceControl.Shutdown()

	conn := dataSourceControl.GetConn()

	// var version string
	// if err = conn.Raw("select version()").Scan(&version).Error; err != nil {
	// 	fmt.Printf("get mysql version failed: %v\n", err)
	// 	return
	// }
	// fmt.Printf("mysql version: %s\n", version)
	ctx := context.Background()

	// Migrate the schema
	conn.AutoMigrate(&Product{})

	// Create
	err = gorm.G[Product](conn).Create(ctx, &Product{Code: "D42", Price: 100})

	// Read
	product, err := gorm.G[Product](conn).Where("id = ?", 1).First(ctx)       // find product with integer primary key
	products, err := gorm.G[Product](conn).Where("code = ?", "D42").Find(ctx) // find product with code D42
	fmt.Printf("product: %v\n", product)
	fmt.Printf("products: %v\n", products)

	// Update - update product's price to 200
	gorm.G[Product](conn).Where("id = ?", product.ID).Update(ctx, "Price", 200)

	// Update - update multiple fields
	gorm.G[Product](conn).Where("id = ?", product.ID).Updates(ctx, Product{Price: 200, Code: "F42"})

	// Delete - delete product
	gorm.G[Product](conn).Where("id = ?", product.ID).Delete(ctx)

	var product1 Product

	// Create
	conn.Create(&Product{Code: "D42", Price: 100})

	// Read
	conn.First(&product1, 1)                 // find product with integer primary key
	conn.First(&product1, "code = ?", "D42") // find product with code D42

	// Update - update product's price to 200
	conn.Model(&product1).Update("Price", 200)
	// Update - update multiple fields
	conn.Model(&product1).Updates(Product{Price: 200, Code: "F42"}) // non-zero fields
	conn.Model(&product1).Updates(map[string]interface{}{"Price": 200, "Code": "F42"})

	// Delete - delete product
	conn.Delete(&product1, 1)
}
