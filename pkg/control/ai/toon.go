package ai

import (
	"github.com/alpkeskin/gotoon"
	"go.uber.org/zap"
)

// ToonEncoder TOON编码器
type ToonEncoder struct {
	logger *zap.SugaredLogger
}

// NewToonEncoder 创建TOON编码器
func NewToonEncoder(toonLogger *zap.SugaredLogger) *ToonEncoder {
	return &ToonEncoder{
		logger: toonLogger,
	}
}

// EncodeToToon 将数据编码为TOON格式
func (encoder *ToonEncoder) EncodeToToon(data interface{}) (string, error) {
	encoded, err := gotoon.Encode(data)
	if err != nil {
		encoder.logger.Errorf("[ai/toon] failed to encode data to TOON format: %v", err)
		return "", err
	}

	return encoded, nil
}

// EncodeToToonWithTabDelimiter 使用制表符分隔符将数据编码为TOON格式
func (encoder *ToonEncoder) EncodeToToonWithTabDelimiter(data interface{}) (string, error) {
	encoded, err := gotoon.Encode(data, gotoon.WithDelimiter("\t"))
	if err != nil {
		encoder.logger.Errorf("[ai/toon] failed to encode data to TOON format with tab delimiter: %v", err)
		return "", err
	}

	return encoded, nil
}

// EncodeToToonWithLengthMarker 使用长度标记将数据编码为TOON格式
func (encoder *ToonEncoder) EncodeToToonWithLengthMarker(data interface{}) (string, error) {
	encoded, err := gotoon.Encode(data, gotoon.WithLengthMarker())
	if err != nil {
		encoder.logger.Errorf("[ai/toon] failed to encode data to TOON format with length marker: %v", err)
		return "", err
	}

	return encoded, nil
}
