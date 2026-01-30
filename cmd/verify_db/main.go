package main

import (
	"fmt"
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

	var bankCount int64
	db.Model(&model.Bank{}).Count(&bankCount)
	fmt.Printf("Banks: %d\n", bankCount)

	var currencyCount int64
	db.Model(&model.Currency{}).Count(&currencyCount)
	fmt.Printf("Currencies: %d\n", currencyCount)
}
