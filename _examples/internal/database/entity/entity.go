package entity

type User struct {
	Uid      int64  `json:"uid" gorm:"column:uid;type:int;primary_key;auto_increment;comment:'自增id'"`
	UserName string `json:"user_name" gorm:"column:user_name;type:varchar(32);not null;comment:'用户名'"`
}

func (User) TableName() string {
	return "t_user"
}

func (u User) String() string {
	return ""
}
