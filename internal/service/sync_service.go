package service

import (
	"encoding/json"
	"fmt"
	"hui_lv_server/config"
	"hui_lv_server/internal/model"
	"hui_lv_server/pkg/db"
	"hui_lv_server/pkg/request"
	"log"
	"strconv"
	"time"
)

// FlexibleExchangeData handles ExchangeItem being object or list
type FlexibleExchangeData struct {
	Data ExchangeItem
}

func (f *FlexibleExchangeData) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &f.Data); err == nil {
		return nil
	}
	var list []ExchangeItem
	if err := json.Unmarshal(data, &list); err == nil {
		if len(list) > 0 {
			f.Data = list[0]
		}
		return nil
	}
	return nil
}

// FlexibleBankData handles BankItem being object or list
type FlexibleBankData struct {
	Data BankItem
}

func (f *FlexibleBankData) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &f.Data); err == nil {
		return nil
	}
	var list []BankItem
	if err := json.Unmarshal(data, &list); err == nil {
		if len(list) > 0 {
			f.Data = list[0]
		}
		return nil
	}
	return nil
}

type SyncService struct{}

// ExchangeItem struct
type ExchangeItem struct {
	From       string `json:"from"`
	FromName   string `json:"from_name"`
	To         string `json:"to"`
	ToName     string `json:"to_name"`
	Exchange   string `json:"exchange"`
	UpdateTime string `json:"updatetime"`
}

// TanshuExchangeResponse structure
type TanshuExchangeResponse struct {
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
	Data FlexibleExchangeData `json:"data"`
}

// BankItem struct
type BankItem struct {
	Time     string `json:"time"`
	Name     string `json:"name"`
	CodeList []struct {
		Code    string `json:"code"`
		HuiIn   string `json:"hui_in"`
		ChaoIn  string `json:"chao_in"`
		HuiOut  string `json:"hui_out"`
		ChaoOut string `json:"chao_out"`
		Zhesuan string `json:"zhesuan"`
		Name    string `json:"name"`
	} `json:"code_list"`
}

// TanshuBankResponse structure
type TanshuBankResponse struct {
	Code int              `json:"code"`
	Msg  string           `json:"msg"`
	Data FlexibleBankData `json:"data"`
}

var MajorCurrencies = []string{
	"USD", "EUR", "HKD", "JPY", "GBP", "AUD", "CAD", "SGD", "CHF", "RUB",
	"KRW", "THB", "NZD", "MOP", "TWD", "PHP", "MYR", "DKK", "SEK", "NOK",
	"TRY", "BRL", "INR", "IDR", "ILS", "ZAR", "SAR", "AED",
}

// Helper to get currency ID by code or name
func (s *SyncService) getCurrencyID(codeOrName string) uint {
	var c model.Currency
	// Try matching Code or Name or BankName
	if err := db.DB.Where("code = ? OR name = ? OR bname = ?", codeOrName, codeOrName, codeOrName).First(&c).Error; err != nil {
		log.Printf("Currency not found: %s", codeOrName)
		return 0
	}
	return c.ID
}

// Helper to get bank ID by code
func (s *SyncService) getBankID(code string) uint {
	var b model.Bank
	if err := db.DB.Where("code = ?", code).First(&b).Error; err != nil {
		log.Printf("Bank not found: %s", code)
		return 0
	}
	return b.ID
}

func (s *SyncService) SyncAll() {
	log.Println("Starting SyncAll...")
	//s.SyncExchangeRates()
	// s.SyncBankRates() // Deprecated: Tanshu API
	s.SyncBankRatesFromKylc()
	log.Println("SyncAll Finished.")
}

func (s *SyncService) SyncExchangeRates() {
	for _, code := range MajorCurrencies {
		url := fmt.Sprintf("%s/exchange/v1/index2?key=%s&from=%s&to=CNY&money=1", config.AppConfig.Tanshu.ApiUrl, config.AppConfig.Tanshu.ApiKey, code)
		var resp TanshuExchangeResponse
		err := request.Get(url, &resp)
		if err != nil {
			log.Printf("Failed to fetch exchange rate for %s: %v", code, err)
			continue
		}

		if resp.Code != 1 {
			log.Printf("API Error for %s: %s url: %s", code, resp.Msg, url)
			continue
		}

		item := resp.Data.Data
		if item.Exchange == "" {
			continue
		}
		rate, _ := strconv.ParseFloat(item.Exchange, 64)

		// Use Today's date YYYY-MM-DD

		today := time.Now().Format("2006-01-02")
		currencyID := s.getCurrencyID(code)
		if currencyID == 0 {
			continue
		}

		record := model.ExchangeRate{
			Date:       today,
			CurrencyID: currencyID,
			Base:       "CNY",
			Rate:       rate,
		}

		// Update or Create
		var count int64
		// Join to check by CurrencyID? Or just trust the ID?
		// Note: The original logic used `currency` string queries. Now we need to query by ID.
		db.DB.Model(&model.ExchangeRate{}).Where("date = ? AND currency_id = ?", today, currencyID).Count(&count)
		if count == 0 {
			db.DB.Create(&record)
		} else {
			db.DB.Model(&model.ExchangeRate{}).Where("date = ? AND currency_id = ?", today, currencyID).Updates(&record)
		}
	}
}

func (s *SyncService) SyncBankRates() {
	var banks []model.Bank
	if err := db.DB.Find(&banks).Error; err != nil {
		log.Printf("Failed to fetch banks: %v", err)
		return
	}

	for _, bank := range banks {
		bankCode := bank.Code
		url := fmt.Sprintf("%s/bank_exchange/v1/index?key=%s&bank_code=%s", config.AppConfig.Tanshu.ApiUrl, config.AppConfig.Tanshu.ApiKey, bankCode)
		var resp TanshuBankResponse
		err := request.Get(url, &resp)
		if err != nil {
			log.Printf("Failed to fetch bank rate for %s: %v", bankCode, err)
			continue
		}

		if resp.Code != 1 {
			log.Printf("API Error for Bank %s: %s", bankCode, resp.Msg)
			continue
		}

		today := time.Now().Format("2006-01-02")
		bankID := s.getBankID(bankCode)
		if bankID == 0 {
			continue
		}

		bankData := resp.Data.Data
		if len(bankData.CodeList) == 0 {
			continue
		}

		for _, item := range bankData.CodeList {
			huiIn, _ := strconv.ParseFloat(item.HuiIn, 64)
			chaoIn, _ := strconv.ParseFloat(item.ChaoIn, 64)
			huiOut, _ := strconv.ParseFloat(item.HuiOut, 64)
			chaoOut, _ := strconv.ParseFloat(item.ChaoOut, 64)
			zhesuan, _ := strconv.ParseFloat(item.Zhesuan, 64)

			if bankCode != "SPDB" {
				huiIn = huiIn / 100
				chaoIn = chaoIn / 100
				huiOut = huiOut / 100
				chaoOut = chaoOut / 100
				zhesuan = zhesuan / 100
			}

			currencyID := s.getCurrencyID(item.Code)
			if currencyID == 0 {
				continue
			}

			record := model.BankRate{
				Date:       today,
				BankID:     bankID,
				CurrencyID: currencyID,
				HuiIn:      huiIn,
				ChaoIn:     chaoIn,
				HuiOut:     huiOut,
				ChaoOut:    chaoOut,
				Zhesuan:    zhesuan,
			}

			// Update or Create
			var count int64
			db.DB.Model(&model.BankRate{}).
				Where("date = ? AND bank_id = ? AND currency_id = ?", today, bankID, currencyID).
				Count(&count)

			if count == 0 {
				db.DB.Create(&record)
			} else {
				db.DB.Model(&model.BankRate{}).
					Where("date = ? AND bank_id = ? AND currency_id = ?", today, bankID, currencyID).
					Updates(&record)
			}
		}
	}
}

// BaiduFinanceResponse struct
type BaiduFinanceResponse struct {
	Result struct {
		RevCode []struct {
			Price string `json:"price"`
		} `json:"revCode"`
	} `json:"Result"`
}

func (s *SyncService) SyncRealTimeRates() {
	log.Println("Starting SyncRealTimeRates...")

	// 1. Get all supported currencies
	var currencies []model.Currency
	if err := db.DB.Find(&currencies).Error; err != nil {
		log.Printf("Failed to fetch currencies: %v", err)
		return
	}

	today := time.Now().Format("2006-01-02")

	for _, c := range currencies {
		if c.Code == "CNY" {
			continue
		}

		// 2. Fetch data from Baidu Finance
		url := fmt.Sprintf("https://finance.pae.baidu.com/api/getrevforeigndata?query=CNY%s&finClientType=pc", c.Code)
		var resp BaiduFinanceResponse
		// Use a simple HTTP get here since our request pkg fits
		err := request.Get(url, &resp)
		if err != nil {
			log.Printf("Failed to fetch real-time rate for %s: %v", c.Code, err)
			continue
		}

		if len(resp.Result.RevCode) == 0 {
			log.Printf("No price data found for %s", c.Code)
			continue
		}

		priceStr := resp.Result.RevCode[0].Price
		rate, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			log.Printf("Failed to parse rate for %s (%s): %v", c.Code, priceStr, err)
			continue
		}

		// 3. Update ExchangeRate
		currencyID := c.ID
		record := model.ExchangeRate{
			Date:       today,
			CurrencyID: currencyID,
			Base:       "CNY",
			Rate:       rate,
		}

		var count int64
		db.DB.Model(&model.ExchangeRate{}).Where("date = ? AND currency_id = ?", today, currencyID).Count(&count)
		if count == 0 {
			db.DB.Create(&record)
		} else {
			db.DB.Model(&model.ExchangeRate{}).Where("date = ? AND currency_id = ?", today, currencyID).Updates(&record)
		}
	}
	log.Println("SyncRealTimeRates Finished.")
}
