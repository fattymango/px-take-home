package model

import "gorm.io/gorm"

type CommonModel struct {
	CreatedAt int64          `gorm:"autoCreateTime:nano;column:created_at" json:"created_at" format:"int64"`
	UpdatedAt int64          `gorm:"autoUpdateTime:nano;column:updated_at" json:"updated_at" format:"int64"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}
