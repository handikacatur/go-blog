package models

import "github.com/jinzhu/gorm"

type Block struct {
	gorm.Model
	PostID    uint
	Type      string
	Component Component
}
