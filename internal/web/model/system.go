package model

import (
	"database/sql"

	"github.com/jianlu8023/go-tools/v2/pkg/helper/json"
)

const (
	systemInitTableName = "t_system_init_status"
)

type SystemInit struct {
	Id     uint         `json:"uid,omitempty" yaml:"uid,omitempty" gorm:"primaryKey;check:id=1"`                          // 主键 确保id 始终是1
	IsInit sql.NullBool `json:"is_init,omitempty" yaml:"is_init,omitempty" gorm:"column:is_init;not null;default:false;"` // 是否已经初始化
}

func (s *SystemInit) MarshalJSON() ([]byte, error) {
	type Alias SystemInit
	aux := &struct {
		*Alias
		IsInit bool `json:"is_init,omitempty" yaml:"is_init,omitempty"`
	}{
		Alias:  (*Alias)(s),
		IsInit: s.IsInit.Bool,
	}
	return json.Marshal(aux)
}

func (s *SystemInit) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (*SystemInit) TableName() string {
	return systemInitTableName
}
