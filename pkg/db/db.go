package db

import (
	"fmt"
	"hui_lv_server/config"
	"hui_lv_server/internal/model"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	// DSN: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
	dbConfig := config.AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.DBName,
	)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Drop PreciousMetal table to ensure schema compatibility (dev only fix for FK)
	if err := DB.Migrator().DropTable(&model.PreciousMetal{}); err != nil {
		log.Printf("Warning: failed to drop precious_metals table: %v", err)
	}

	// Auto Migrate
	err = DB.AutoMigrate(
		&model.Bank{}, 
		&model.Currency{}, 
		&model.ExchangeRate{}, 
		&model.BankRate{}, 
		&model.UserSubscription{}, 
		&model.BankDeposit{}, 
		&model.Admin{},
		&model.PreciousMetal{},
	)
	if err != nil {
		log.Fatal("failed to migrate database")
	}
}
