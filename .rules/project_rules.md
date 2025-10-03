注释格式
// QueryFileExist 查询文件是否存在
// @description 根据文件查询条件查询文件是否存在
// @param query *model.FileInfo 文件查询条件，非空字段将作为查询条件
// @return bool 文件是否存在
// @return error 错误信息

web模块/handler包
做请求接收，绑定参数和参数验证，可参考现有handler包中的其他模块如何做请求接收和参数绑定

web模块/service包
做业务逻辑处理，可参考现有service包中的其他模块如何做逻辑操作

web模块/request包
定义接口请求参数结构体和参数验证方法生成，可参考现有request包中的结构体和方法

web模块/response包
定义接口响应值的具体结构体并做一定字段转换和json序列化，可参考现有response包中的结构体和方法

web模块/model包
定义数据库模型结构体，可参考现有model包中的结构体，需要实现有String()string方法和TableName()string方法 ，
String()string方法用于打印模型结构体信息，TableName()string方法用于指定数据库表名

web模块/mapper包
定义数据库模型结构体和接口请求参数结构体之间的映射关系，可参考现有mapper包中的结构体和方法
首先进行m.db是否未空的判断，若为空则返回错误；其次如果是insert update delete使用transaction进行

附加提示：
若不清楚有什么方法可调用，请直接通过import的地方查看项目中具体的文件，进行学习后编写，若真的没有需要调用的方法，并觉得后续可能会用到，可以添加到对应的文件中，
若觉得只使用一次，可在调用的地方直接实现，后续我考虑是否加入对应的文件。

