package models

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique_index;not null" form:"username"`
	Password string `form:"password"`
	Posts    []Post `gorm:"foreignkey:UserID"`
}
