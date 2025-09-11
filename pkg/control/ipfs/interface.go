package ipfs

// // ControlInterface 定义IPFS控制器的接口
// //
// // @description 该接口定义了IPFS控制模块需要实现的核心功能
// // @interface
// type ControlInterface interface {
// 	// UploadFile 上传文件到IPFS
// 	//
// 	// @param filePath string 要上传的文件路径
// 	// @return string 文件在IPFS中的CID
// 	// @return error 上传过程中的错误
// 	UploadFile(filePath string) (string, error)
//
// 	// DownloadFile 从IPFS下载文件
// 	//
// 	// @param cidStr string IPFS中的文件CID
// 	// @param outputPath string 下载后的文件保存路径
// 	// @return error 下载过程中的错误
// 	DownloadFile(cidStr string, outputPath string) error
//
// 	// GetFileContent 获取IPFS文件内容
// 	//
// 	// @param cidStr string IPFS中的文件CID
// 	// @return []byte 文件内容
// 	// @return error 获取过程中的错误
// 	GetFileContent(cidStr string) ([]byte, error)
// }
