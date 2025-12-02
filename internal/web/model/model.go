package model

import (
	"time"

	"github.com/jianlu8023/golang-example/pkg/snowflake"
	"gorm.io/gorm"
)

type Model struct {
	RawUid         int64     `json:"raw_uid" yaml:"raw_uid" gorm:"column:raw_uid;type:int;comment:'行id';"`
	CreateDateTime time.Time `json:"create_date_time" yaml:"create_date_time" gorm:"column:create_date_time;type:datetime;comment:'记录创建时间';autoCreateTime:true;"`
	UpdateDateTime time.Time `json:"update_date_time" yaml:"update_date_time" gorm:"column:update_date_time;type:datetime;comment:'记录更新时间';autoUpdateTime:true;"`
}

func (m *Model) BeforeSave(db *gorm.DB) error {
	if m.RawUid == 0 {
		id, err := snowflake.GenID()
		if err != nil {
			return err
		}
		m.RawUid = int64(id)
	}

	// if m.CreateDateTime.IsZero() {
	// 	m.CreateDateTime = time.Now()
	// }
	// if m.UpdateDateTime.IsZero() {
	// 	m.UpdateDateTime = time.Now()
	// }
	return nil
}

// func (m *Model) BeforeUpdate(db *gorm.DB) error {
// 	m.UpdateDateTime = time.Now()
// 	return nil
// }

func (m *Model) BeforeCreate(db *gorm.DB) error {
	if m.RawUid == 0 {
		id, err := snowflake.GenID()
		if err != nil {
			return err
		}
		m.RawUid = int64(id)
	}
	// if m.CreateDateTime.IsZero() {
	// 	m.CreateDateTime = time.Now()
	// }
	// if m.UpdateDateTime.IsZero() {
	// 	m.UpdateDateTime = time.Now()
	// }
	return nil
}

// func (m *Model) MarshalJSON() ([]byte, error) {
// 	type Alias Model
// 	aux := struct {
// 		*Alias
// 	}{
// 		Alias: (*Alias)(m),
// 	}
// 	return json.Marshal(aux)
// }
//
// func (m *Model) GoString() string {
// 	return m.String()
// }
//
// func (m *Model) String() string {
// 	marshalString, _ := json.MarshalString(m)
// 	return marshalString
// }
