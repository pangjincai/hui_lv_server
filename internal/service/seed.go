package service

import (
	"hui_lv_server/internal/consts"
	"hui_lv_server/internal/model"
	"hui_lv_server/pkg/db"
	"hui_lv_server/pkg/utils"
	"log"
)

func Seed() {
	log.Println("Seeding Banks and Currencies...")
	/*seedBanks()
	seedBanks()
	seedCurrencies()
	seedAdmin()*/
	log.Println("Seeding setup finished.")
}

func seedAdmin() {
	var admin model.Admin
	result := db.DB.Where("username = ?", "admin").First(&admin)
	
	if result.Error != nil {
		// Create new
		salt := utils.GenerateSalt()
		password := utils.HashPassword("123123", salt)
		db.DB.Create(&model.Admin{Username: "admin", Password: password, Salt: salt})
		log.Println("Seeded Admin: admin")
	} else if admin.Salt == "" {
		// Update existing legacy admin
		salt := utils.GenerateSalt()
		password := utils.HashPassword("123123", salt)
		admin.Salt = salt
		admin.Password = password
		db.DB.Save(&admin)
		log.Println("Updated Admin to use Salt+MD5")
	}
}

func seedBanks() {
	bankNames := map[string]string{
		"ICBC":     "工商银行",
		"BOC":      "中国银行",
		"ABCHINA":  "农业银行",
		"BANKCOMM": "交通银行",
		"CCB":      "建设银行",
		"CMBCHINA": "招商银行",
		"CEBBANK":  "中国光大银行",
		"SPDB":     "上海浦东发展银行",
		"CIB":      "兴业银行",
		"ECITIC":   "中信银行",
		"PSBC":     "邮政储蓄银行",
		"CMBC":     "民生银行",
	}

	for code, name := range bankNames {
		var count int64
		db.DB.Model(&model.Bank{}).Where("code = ?", code).Count(&count)
		if count == 0 {
			db.DB.Create(&model.Bank{Code: code, Name: name})
			log.Printf("Seeded Bank: %s - %s", code, name)
		} else {
			// Update name if changed
			db.DB.Model(&model.Bank{}).Where("code = ?", code).Update("name", name)
		}
	}
}

func seedCurrencies() {
	// Priority list for sorting
	priority := map[string]int{
		"USD": 1,
		"HKD": 2,
		"JPY": 3,
		"EUR": 4,
		"GBP": 5,
		"AUD": 6,
		"CAD": 7,
		"SGD": 8,
		"CHF": 9,
		"MOP": 10,
		"RUB": 11,
		"KRW": 12,
		"THB": 13,
		"TWD": 14,
	}

	for code, name := range consts.CurrencyNames {
		sortOrder := 100 // Default sort order
		if p, ok := priority[code]; ok {
			sortOrder = p
		}

		var count int64
		db.DB.Model(&model.Currency{}).Where("code = ?", code).Count(&count)
		if count == 0 {
			db.DB.Create(&model.Currency{Code: code, Name: name, Sort: sortOrder})
			log.Printf("Seeded Currency: %s - %s (Sort: %d)", code, name, sortOrder)
		} else {
			// Update name and sort if changed
			db.DB.Model(&model.Currency{}).Where("code = ?", code).Updates(map[string]interface{}{
				"name": name,
				"sort": sortOrder,
			})
		}
	}
}
