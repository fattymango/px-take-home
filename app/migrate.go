package app

import (
	"github.com/fattymango/px-take-home/config"
	"github.com/fattymango/px-take-home/model"
	"github.com/fattymango/px-take-home/pkg/db"
)

func Migrate(cfg *config.Config, db *db.DB) error {

	db.AutoMigrate(&model.Task{})
	return nil
}
