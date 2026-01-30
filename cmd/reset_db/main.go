package main

import (
	"hui_lv_server/internal/model"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:123123@tcp(localhost:3306)/huilv?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	// Drop tables in reverse dependency order
	log.Println("Dropping tables...")
	// Ignore errors safely
	db.Migrator().DropTable(&model.BankRate{})
	db.Migrator().DropTable(&model.BankDeposit{})
	db.Migrator().DropTable(&model.ExchangeRate{})
	db.Migrator().DropTable(&model.UserSubscription{})
	db.Migrator().DropTable(&model.Bank{})
	db.Migrator().DropTable(&model.Currency{})
	log.Println("Tables dropped.")
}
