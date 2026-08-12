package webrtc

import (
	"github.com/pion/webrtc/v4"
)

// MediaType 定义媒体类型
type MediaType string

const (
	// MediaTypeAudio 音频类型
	MediaTypeAudio MediaType = "audio"
	// MediaTypeVideo 视频类型
	MediaTypeVideo MediaType = "video"
	// MediaTypeData 数据类型
	MediaTypeData MediaType = "data"
	// MediaTypeUnknown 未知类型
	MediaTypeUnknown MediaType = "unknown"
)

// MediaTrack 表示一个媒体轨道
//
// @description 表示一个媒体轨道，封装本地静态RTP轨道及其对应的RTPSender，并提供媒体类型与标签信息
// @struct
type MediaTrack struct {
	// ID 媒体轨道唯一标识
	ID string
	// Type 媒体类型（audio/video/data/unknown）
	Type MediaType
	// Label 媒体轨道标签
	Label string
	// Kind 媒体轨道种类
	Kind string
	// Track 本地静态RTP轨道
	Track *webrtc.TrackLocalStaticRTP
	// Sender 与该轨道关联的RTPSender，用于从底层连接移除该轨道
	Sender *webrtc.RTPSender
}

// getMediaType 根据类型字符串获取MediaType枚举值
//
// @description 根据输入字符串返回对应的 MediaType 枚举值，未知类型返回 MediaTypeUnknown 而非默认 data，避免误判
// @param kind string 媒体类型字符串
// @return MediaType 对应的 MediaType 枚举值
func getMediaType(kind string) MediaType {
	switch kind {
	case "audio":
		return MediaTypeAudio
	case "video":
		return MediaTypeVideo
	case "data":
		return MediaTypeData
	default:
		return MediaTypeUnknown
	}
}
