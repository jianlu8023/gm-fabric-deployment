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
)

// MediaTrack 表示一个媒体轨道
type MediaTrack struct {
	ID       string
	Type     MediaType
	Label    string
	Kind     string
	Track    *webrtc.TrackLocalStaticRTP
	Receiver *webrtc.RTPReceiver
}

// getMediaType 根据类型字符串获取MediaType枚举值
func getMediaType(kind string) MediaType {
	switch kind {
	case "audio":
		return MediaTypeAudio
	case "video":
		return MediaTypeVideo
	case "data":
		return MediaTypeData
	default:
		return MediaTypeData
	}
}
