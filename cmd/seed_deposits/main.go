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

	seedData(db)
}

func seedData(db *gorm.DB) {
	// Helper to get ID
	getBankID := func(code string) uint {
		var b model.Bank
		db.Where("code = ?", code).First(&b)
		return b.ID
	}
	getCurrencyID := func(code string) uint {
		var c model.Currency
		db.Where("code = ?", code).First(&c)
		return c.ID
	}

	icbcID := getBankID("ICBC")
	bocID := getBankID("BOC")
	cmbID := getBankID("CMBCHINA")

	usdID := getCurrencyID("USD")
	hkdID := getCurrencyID("HKD")

	if icbcID == 0 || bocID == 0 || cmbID == 0 || usdID == 0 || hkdID == 0 {
		log.Println("Missing banks or currencies, skipping seed.")
		return
	}

	deposits := []model.BankDeposit{
		{
			BankID:     icbcID,
			CurrencyID: usdID,
			Name:       "工行个人外币存款 (ICBC Personal)",
			Time:       "1个月 (1 Month)",
			Rate:       "0.1%",
			Tips:       "起存金额 5000 USD",
		},
		{
			BankID:     icbcID,
			CurrencyID: usdID,
			Name:       "工行美元定期 (ICBC Fixed)",
			Time:       "6个月 (6 Months)",
			Rate:       "0.5%",
			Tips:       "",
		},
		{
			BankID:     bocID,
			CurrencyID: usdID,
			Name:       "中行美元优享 (BOC Premium)",
			Time:       "1年 (1 Year)",
			Rate:       "1.2%",
			Tips:       "新客户专享 (New customers only)",
		},
		{
			BankID:     cmbID,
			CurrencyID: hkdID,
			Name:       "招行港币储蓄 (CMB HKD Savings)",
			Time:       "活期 (Current)",
			Rate:       "0.01%",
			Tips:       "",
		},
	}

	for _, d := range deposits {
		db.Create(&d)
		fmt.Printf("Created deposit: %s\n", d.Name)
	}
}
