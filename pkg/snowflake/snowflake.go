package snowflake

import (
	"time"

	"github.com/sony/sonyflake"
)

var sf *sonyflake.Sonyflake

func init() {
	sf = sonyflake.NewSonyflake(sonyflake.Settings{
		StartTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	})
	if sf == nil {
		panic("sonyflake init failed")
	}
}

func GenID() (uint64, error) {
	return sf.NextID()
}
