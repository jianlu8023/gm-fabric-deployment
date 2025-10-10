# 注释格式

* 方法和函数

```text
// InsertOneWithCheck 插入一条Docker镜像信息，如果存在（包括已逻辑删除的）则返回datasource.ErrAlreadyExists
//
// @description 在事务中插入一条Docker镜像信息，如果数据库中已存在相同名称和位置的镜像（包括已逻辑删除的），则返回错误
// @param imageInfo *model.DockerImage Docker镜像信息
// @return error 错误信息
func (m *DockerImageMapper) InsertOneWithCheck(imageInfo *model.DockerImage) error {}
```

* 结构体

```text
// SystemInit 系统初始化状态模型
//
// @description 存储系统的初始化状态信息
// @struct
type SystemInit struct {
    Id     uint         `json:"uid,omitempty" yaml:"uid,omitempty" gorm:"primaryKey;check:id=1"`                          // 主键 确保id 始终是1
    IsInit sql.NullBool `json:"is_init,omitempty" yaml:"is_init,omitempty" gorm:"column:is_init;not null;default:false;"` // 是否已经初始化
}
```

* 接口

```text
// Libp2pNodeServiceInterface 节点服务接口
// 
// @description 定义节点服务的接口
// @interface
type Libp2pNodeServiceInterface interface {
    // Libp2pNodeList 获取节点列表的处理函数
    // 
    // @description 处理获取节点列表的HTTP请求，验证参数并调用服务层获取节点列表
    // @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
	Libp2pNodeList(ctx *gin.Context, req *request.Libp2pNodeListRequest)
	// Libp2pNodeMyself 获取本机节点信息的处理函数
    //
    // @description 处理获取本机节点信息的HTTP请求，验证参数并调用服务层获取本机节点信息
    // @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
	Libp2pNodeMyself(ctx *gin.Context, req *request.Libp2pNodeMyselfRequest)
}
```

* web部分的handler
```text
// RegisterUserHandler 用户注册处理函数
//
// @description 处理用户注册的HTTP请求，从请求体中解析用户注册信息，调用service方法处理业务逻辑，并通过HTTP响应返回结果
// @method POST
// @url /user/register
// @param ctx *gin.Context Gin上下文，包含HTTP请求和响应对象
func (h *UserHandler) RegisterUserHandler(ctx *gin.Context) {}
```

* 在方法内必要的地方添加注释

# web模块 internal/web/

1. web模块/handler包
   做请求接收，绑定参数和参数验证，可参考现有handler包中的其他模块如何做请求接收和参数绑定

2. web模块/service包
   做业务逻辑处理，可参考现有service包中的其他模块如何做逻辑操作

3. web模块/request包
   定义接口请求参数结构体和参数验证方法生成，可参考现有request包中的结构体和方法

4. web模块/response包
   定义接口响应值的具体结构体并做一定字段转换和json序列化，可参考现有response包中的结构体和方法

5. web模块/model包
   定义数据库模型结构体，可参考现有model包中的结构体，
   需要实现有String()string方法和TableName()string方法，
   String()string方法用于打印模型结构体信息，TableName()string方法用于指定数据库表名

6. web模块/mapper包
   定义数据库模型结构体和接口请求参数结构体之间的映射关系，
   可参考现有mapper包中的结构体和方法
   首先进行m.db是否未空的判断，若为空则返回错误；
   其次如果是insert update delete使用transaction进行

# 控制器 pkg/control/

1. control模块
   定义一个特定的控制器，
   可参考现有控制器，
   最好参考http控制器，datasource控制器，config控制器，tracer控制器，logger控制器，server控制器
   首先检查config控制器的config是否已经定义了对应的结构体；
   其次针对控制器创建对应的文件夹；
   第三必须要有默认配置且enabled=false；
   第四必须实现server控制的iface.go的接口；

# 附加提示

1. 若不清楚有什么方法可调用，请直接通过import的地方查看项目中具体的文件，
   进行学习后编写，若真的没有需要调用的方法，并觉得后续可能会用到，可以添加到对应的文件中，
   若觉得只使用一次，可在调用的地方直接实现，后续我考虑是否加入对应的文件。

# 重要提示

1. 不要主动执行go mod tidy 命令，可以告知需要什么依赖，我自行添加并执行
2. 可以调用golangci-lint相关命令，但是执行前先进行安装 目前项目中的.golangci.toml文件已配置好v2版本
   安装命令 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
3. 如果需要运行程序server程序 相关命令 ./server -c configs/server.yaml -t win -d -v
   启动前检查configs/server-win.yaml的配置，win需要运行server.exe linux需要运行server.bin
   需要检查程序代码参考cmd/server/，
   打包命令参考makefile中的信息 linux直接make build即可
   win命令go build -tags=jsoniter -trimpath -ldflags="-s -w" -o server.exe cmd/server/server.go
4. 如果需要运行程序client程序 相关命令 ./client -c configs/client.yaml -t win -d -v
   启动前检查configs/client-win.yaml的配置，win需要运行client.exe linux需要运行client.bin
   需要检查程序代码参考cmd/client/，
   打包命令参考makefile中的信息 linux直接make build即可
   win命令go build -tags=jsoniter -trimpath -ldflags="-s -w" -o client.exe cmd/client/client.go
5. 所有.go文件最后一行必须添加一行空行
6. 占位符

# 核心提示

1. 修改或添加的代码必须符合golang规范(针对后缀是.go文件)
