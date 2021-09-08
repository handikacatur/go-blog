package models

import (
	"github.com/jinzhu/gorm"
)

type Post struct {
	gorm.Model
	Title    string `form:"title"`
	Subtitle string `form:"subtitle"`
	Cover    string `form:"cover"`
	Data     string `form:"data"`
	UserID   uint
}
