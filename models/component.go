package models

import "github.com/jinzhu/gorm"

type Component struct {
	gorm.Model
	Text    string
	File    string
	Caption string
	Item    string
}
