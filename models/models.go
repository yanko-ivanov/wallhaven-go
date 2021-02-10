package models

import (
	"gorm.io/gorm"
)

type Wallpaper struct {
	gorm.Model
	Url       string
	Path      string
	ThumbPath string
}
