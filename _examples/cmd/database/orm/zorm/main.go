package main

import (
	"context"
	"log"

	_ "gitee.com/chunanyong/dm"
	"gitee.com/chunanyong/zorm"
	_ "github.com/go-sql-driver/mysql"
)

// https://zorm.cn/post/doc/index.html

// https://gitee.com/chunanyong/zorm/tree/v1.7.6

func main() {
	dbConfig := &zorm.DataSourceConfig{
		DSN:        "root:123456@tcp(127.0.0.1:3306)/basic?charset=utf8&parseTime=true&loc=Local",
		DriverName: "mysql",
		Dialect:    "mysql",
	}
	dao, err := zorm.NewDBDao(dbConfig)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer func() {
		if err := dao.CloseDB(); err != nil {
			log.Fatalln(err)
		}
	}()
	mctx := context.Background()
	mctx, err = dao.BindContextDBConnection(mctx)
	if err != nil {
		log.Fatalln("mysql绑定ctx失败 ", err)
		return
	}
	finder := zorm.NewFinder().Append("select @@version")
	var mversion string
	row, err := zorm.QueryRow(mctx, finder, &mversion)
	if err != nil {
		log.Fatalln(err)
		return
	}
	if row {
		log.Println(mversion)
	}

	log.Println("starting dameng query")

	damengConfig := &zorm.DataSourceConfig{
		DSN:        "dm://SYSDBA:SYSDBA001@127.0.0.1:5236/basic",
		DriverName: "dm",
		Dialect:    "dm",
	}

	dameng, err := zorm.NewDBDao(damengConfig)
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer func() {
		if err := dameng.CloseDB(); err != nil {
			log.Fatalln(err)
		}
	}()

	dctx := context.Background()
	dctx, err = dameng.BindContextDBConnection(dctx)
	if err != nil {
		log.Fatalln("达梦绑定ctx失败 ", err)
		return
	}

	dfinder := zorm.NewFinder().Append("SELECT * FROM V$VERSION")
	var version []interface{}
	err = zorm.Query(dctx, dfinder, &version, nil)
	if err != nil {
		log.Fatalln("query error ", err)
		return
	}
	log.Println(version)

}

// // CustomDMText 实现ICustomDriverValueConver接口,扩展自定义类型,例如 达梦数据库TEXT类型,映射出来的是dm.DmClob类型,无法使用string类型直接接收
// type CustomDMText struct{}

// // GetDriverValue 根据数据库列类型,返回driver.Value的实例,struct属性类型
// // map接收或者字段不存在,无法获取到structFieldType,会传入nil
// func (dmtext CustomDMText) GetDriverValue(ctx context.Context, columnType *sql.ColumnType, structFieldType *reflect.Type) (driver.Value, error) {
// 	// 如果需要使用structFieldType,需要先判断是否为nil
// 	// if structFieldType != nil {
// 	// }

// 	return &dm.DmClob{}, nil
// }

// // ConverDriverValue 数据库列类型,GetDriverValue返回的driver.Value的临时接收值,struct属性类型
// // map接收或者字段不存在,无法获取到structFieldType,会传入nil
// // 返回符合接收类型值的指针,指针,指针!!!!
// func (dmtext CustomDMText) ConverDriverValue(ctx context.Context, columnType *sql.ColumnType, tempDriverValue driver.Value, structFieldType *reflect.Type) (interface{}, error) {
// 	// 如果需要使用structFieldType,需要先判断是否为nil
// 	// if structFieldType != nil {
// 	// }

// 	// 类型转换
// 	dmClob, isok := tempDriverValue.(*dm.DmClob)
// 	if !isok {
// 		return tempDriverValue, errors.New("->ConverDriverValue-->转换至*dm.DmClob类型失败")
// 	}
// 	if dmClob == nil || !dmClob.Valid {
// 		return new(string), nil
// 	}
// 	// 获取长度
// 	dmlen, errLength := dmClob.GetLength()
// 	if errLength != nil {
// 		return dmClob, errLength
// 	}

// 	// int64转成int类型
// 	strInt64 := strconv.FormatInt(dmlen, 10)
// 	dmlenInt, errAtoi := strconv.Atoi(strInt64)
// 	if errAtoi != nil {
// 		return dmClob, errAtoi
// 	}

// 	// 读取字符串
// 	str, errReadString := dmClob.ReadString(1, dmlenInt)

// 	// 处理空字符串或NULL造成的EOF错误
// 	if errReadString == io.EOF {
// 		return new(string), nil
// 	}

// 	return &str, errReadString
// }

// // RegisterCustomDriverValueConver 注册自定义的字段处理逻辑,用于驱动无法直接转换的场景,例如达梦的 TEXT 无法直接转化成 string
// // 一般是放到init方法里进行注册
// func init() {
// 	zorm.RegisterCustomDriverValueConver("dm.TEXT", CustomDMText{})

// }
