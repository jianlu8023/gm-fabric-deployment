# golang-web-module-guide

这是一个专门用于生成golang web模块的技能，它可以帮助你快速生成各种golang web模块，包括但不限于路由、控制器、服务、模型等。

## 使用方法

1. 生成到internal/web/文件夹下
2. 按照文件夹用途放置到不同的文件夹下
3. internal/web/下如果没有xxx/xxx.go(如handler/handler.go)，则自动生成,否则不生成
4. 注释需要参考example_web文件夹下的注释规范

## references

- [golang-gin](https://gin-gonic.com/zh-cn/docs/quickstart/)
- [golang-gorm](https://gorm.io/zh_CN/docs/index.html)
- [golang-web-module-example](../asserts/example_web)
