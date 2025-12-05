package response

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/sqlnull"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"github.com/jinzhu/copier"
)

// DockerImageListResponse Docker镜像列表响应结构体
//
// @description 用于返回Docker镜像列表的响应数据
// @struct
type DockerImageListResponse struct {
	ImageName           []string          `json:"image_name" yaml:"image_name"`                         // 镜像名称
	ImageCreated        time.Time         `json:"image_created" yaml:"image_created"`                   // 镜像创建时间
	ImageLabels         map[string]string `json:"image_labels" yaml:"image_labels"`                     // 镜像标签
	ImageId             string            `json:"image_id" yaml:"image_id"`                             // 镜像id
	ImageLocationPeerId string            `json:"image_location_peer_id" yaml:"image_location_peer_id"` // 镜像所在peer
	IsDelete            sql.NullBool      `json:"is_delete" yaml:"is_delete"`                           // 是否删除
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 自定义DockerImageListResponse结构体的JSON序列化逻辑，处理时间格式和sql.NullBool类型
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp DockerImageListResponse) MarshalJSON() ([]byte, error) {
	type Alias DockerImageListResponse
	aux := struct {
		*Alias
		IsDelete     *bool  `json:"is_delete"`
		ImageCreated string `json:"image_created"`
	}{
		Alias:        (*Alias)(&resp),
		IsDelete:     sqlnull.NullBoolPtr(resp.IsDelete),
		ImageCreated: humantime.HumanTime(resp.ImageCreated, "unknown"),
	}
	return json.Marshal(aux)
}

// String 将Docker镜像列表响应转换为字符串表示
//
// @description 将DockerImageListResponse结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp DockerImageListResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewDockerImageListResponse 创建Docker镜像列表响应对象
//
// @description 将数据库模型转换为API响应对象，并创建分页信息
// @param page dbpage.Info[model.DockerImage] 数据库查询得到的分页镜像数据
// @return *dbpage.Info[DockerImageListResponse] 转换后的分页响应数据
// @return error 转换过程中的错误信息
func NewDockerImageListResponse(page dbpage.Info[model.DockerImage]) (*dbpage.Info[DockerImageListResponse], error) {
	var convert []DockerImageListResponse

	var src1 int64 = 0
	// src2 := []string{""}
	err := copier.CopyWithOption(&convert, page.GetRecords(), copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
		Converters: []copier.TypeConverter{
			{
				SrcType: src1,
				DstType: time.Time{},
				Fn: func(src interface{}) (dst interface{}, err error) {
					if src == nil {
						return nil, nil
					}
					timeInt, ok := src.(int64)
					if !ok {
						return nil, nil
					}
					local, err := humantime.ParseTimeLocal(fmt.Sprintf("%v", timeInt))
					if err != nil {
						return nil, err
					}
					return local, nil
				},
			},
			// {
			// 	SrcType: src2,
			// 	DstType: copier.String,
			// 	Fn: func(src interface{}) (dst interface{}, err error) {
			// 		if src == nil {
			// 			return nil, nil
			// 		}
			// 		imageNameSrc, ok := src.([]string)
			// 		if !ok {
			// 			return nil, nil
			// 		}
			// 		imageNameDst, err := json.MarshalString(imageNameSrc)
			// 		if err != nil {
			// 			return nil, err
			// 		}
			// 		return imageNameDst, nil
			// 	},
			// },
		},
	})
	if err != nil {
		return nil, err
	}
	return dbpage.NewPageInfo[DockerImageListResponse](
		page.GetPageNo(),
		page.GetPageSize(),
		page.GetCount(),
		page.GetMaxPage(),
		convert,
	), nil
}

// DockerImagePullResponse Docker镜像拉取响应结构体
//
// @description 用于返回Docker镜像拉取操作的响应数据
// @struct
type DockerImagePullResponse struct {
	Msg string `json:"msg" yaml:"msg"` // 操作结果消息
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 自定义DockerImagePullResponse结构体的JSON序列化逻辑
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp DockerImagePullResponse) MarshalJSON() ([]byte, error) {
	type Alias DockerImagePullResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// String 将Docker镜像拉取响应转换为字符串表示
//
// @description 将DockerImagePullResponse结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp DockerImagePullResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewDockerImagePullResponse 创建Docker镜像拉取响应对象
//
// @description 创建一个新的Docker镜像拉取响应对象
// @param msg string 操作结果消息
// @return *DockerImagePullResponse 镜像拉取响应对象
func NewDockerImagePullResponse(msg string) *DockerImagePullResponse {
	return &DockerImagePullResponse{
		Msg: msg,
	}
}
