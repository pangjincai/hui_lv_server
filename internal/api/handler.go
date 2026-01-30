package api

import (
	"hui_lv_server/config"
	"hui_lv_server/internal/model"
	"hui_lv_server/pkg/db"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// GetRates returns the latest exchange rates
func (h *Handler) GetRates(c *gin.Context) {
	var rates []model.ExchangeRate
	// Get rates for today
	today := time.Now().Format("2006-01-02")
	
	// Join with Currency to sort by Sort field
	db.DB.Preload("Currency").
		Joins("JOIN currencies ON currencies.id = exchange_rates.currency_id").
		Where("exchange_rates.date = ?", today).
		Order("currencies.sort ASC").
		Find(&rates)

	// If empty, maybe sync didn't run? Return previous day or empty.
	if len(rates) == 0 {
		// Fallback to latest available
		var last model.ExchangeRate
		db.DB.Last(&last)
		if last.Date != "" {
			db.DB.Preload("Currency").
				Joins("JOIN currencies ON currencies.id = exchange_rates.currency_id").
				Where("exchange_rates.date = ?", last.Date).
				Order("currencies.sort ASC").
				Find(&rates)
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": rates})
}

// BankRateResponse wraps model.BankRate with formatted price fields
type BankRateResponse struct {
	ID        uint           `json:"ID"`
	CreatedAt model.LocalTime `json:"CreatedAt"`
	UpdatedAt model.LocalTime `json:"UpdatedAt"`
	Date      string         `json:"Date"`
	BankID    uint           `json:"BankID"`
	Bank      model.Bank     `json:"Bank"`
	CurrencyID uint          `json:"CurrencyID"`
	Currency  model.Currency `json:"Currency"`
	HuiIn     interface{}    `json:"HuiIn"`
	ChaoIn    interface{}    `json:"ChaoIn"`
	HuiOut    interface{}    `json:"HuiOut"`
	ChaoOut   interface{}    `json:"ChaoOut"`
	Zhesuan   interface{}    `json:"Zhesuan"`
}

// GetBankRates returns rates for a specific bank or all banks
func (h *Handler) GetBankRates(c *gin.Context) {
	bankCode := c.Query("bank_code")
	currency := c.Query("currency") // Optional filter by currency code

	query := db.DB.Model(&model.BankRate{}).Preload("Bank").Preload("Currency")
	// Use Joins to filter by Bank Code or Currency Code
	query = query.Joins("Bank").Joins("Currency")

	// Default to today or latest
	today := time.Now().Format("2006-01-02")
	var count int64
	db.DB.Model(&model.BankRate{}).Where("date = ?", today).Count(&count)
	if count > 0 {
		query = query.Where("bank_rates.date = ?", today)
	} else {
		// Fallback to latest available date
		var last model.BankRate
		db.DB.Last(&last)
		if last.Date != "" {
			query = query.Where("bank_rates.date = ?", last.Date)
		}
	}
	// Filter out rates where any value is > 0 (OR condition)
	query = query.Where("hui_in > 0 OR chao_in > 0 OR hui_out > 0 OR chao_out > 0")

	if bankCode != "" {
		query = query.Where("Bank.code = ?", bankCode)
	}
	if currency != "" {
		query = query.Where("Currency.code = ?", currency)
	}

	var rates []model.BankRate
	// Sort Logic:
	// If viewing a specific bank (bankCode != ""), sort by Currency.Sort ASC (common currencies first)
	// If viewing comparison (currency != ""), sort by HuiOut ASC (best selling rate first)
	if bankCode != "" {
		query.Order("Currency.sort ASC").Find(&rates)
	} else {
		// Note: 0 values are excluded by previous Where clause, so we get valid lowest selling rates first.
		query.Order("hui_out ASC").Find(&rates)
	}

	// Transform to Response
	var resp []BankRateResponse
	for _, r := range rates {
		item := BankRateResponse{
			ID:        r.ID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
			Date:      r.Date,
			BankID:    r.BankID,
			Bank:      r.Bank,
			CurrencyID: r.CurrencyID,
			Currency:  r.Currency,
			HuiIn:     formatRate(r.HuiIn),
			ChaoIn:    formatRate(r.ChaoIn),
			HuiOut:    formatRate(r.HuiOut),
			ChaoOut:   formatRate(r.ChaoOut),
			Zhesuan:   formatRate(r.Zhesuan),
		}
		resp = append(resp, item)
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": resp})
}

func formatRate(val float64) interface{} {
	if val <= 0 {
		return "--"
	}
	// Multiply by 100 and format to 4 decimal places
	return strconv.FormatFloat(val*100, 'f', 4, 64)
}

// GetHistory returns historical data for charts
func (h *Handler) GetHistory(c *gin.Context) {
	currency := c.Query("currency")
	days := 30 // Default 30 days

	if currency == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "currency is required"})
		return
	}

	var rates []model.ExchangeRate
	// Limit to last 30 days
	db.DB.Preload("Currency").Joins("Currency").
		Where("Currency.code = ?", currency).
		Order("date asc").
		Limit(days).
		Find(&rates)

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": rates})
}

// SubscribeRequest payload
type SubscribeRequest struct {
	OpenID     string  `json:"openid"`
	Currency   string  `json:"currency"`
	Threshold  float64 `json:"threshold"`
	NotifyType string  `json:"notify_type"`
}

// Subscribe adds or updates a user subscription
func (h *Handler) Subscribe(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}

	sub := model.UserSubscription{
		OpenID:     req.OpenID,
		Currency:   req.Currency,
		Threshold:  req.Threshold,
		NotifyType: req.NotifyType,
	}

	var count int64
	db.DB.Model(&model.UserSubscription{}).Where("open_id = ? AND currency = ?", req.OpenID, req.Currency).Count(&count)
	if count > 0 {
		db.DB.Model(&model.UserSubscription{}).Where("open_id = ? AND currency = ?", req.OpenID, req.Currency).Updates(&sub)
	} else {
		db.DB.Create(&sub)
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Success"})
}

// GetSubscriptions returns user subscriptions
func (h *Handler) GetSubscriptions(c *gin.Context) {
	openid := c.Query("openid")
	if openid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "openid is required"})
		return
	}

	var subs []model.UserSubscription
	db.DB.Where("open_id = ?", openid).Find(&subs)

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": subs})
}

// DepositProduct structure
type DepositProduct struct {
	Name         string `json:"name"`
	Time         string `json:"time"`
	BankName     string `json:"bank_name"`
	CurrencyName string `json:"currency_name"`
	CurrencyCode string `json:"currency_code"`
}

// GetDeposits returns foreign currency deposit products
func (h *Handler) GetDeposits(c *gin.Context) {
	var deposits []model.BankDeposit

	// Find all deposits, preload relations
	result := db.DB.Preload("Bank").Preload("Currency").Find(&deposits)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 0,
			"msg":  "Failed to fetch deposits",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": deposits,
	})
}

// GetPreciousMetals returns gold/silver prices
func (h *Handler) GetPreciousMetals(c *gin.Context) {
	var metals []model.PreciousMetal
	today := time.Now().Format("2006-01-02")

	// Get today's data with Preload Currency
	db.DB.Preload("Currency").Where("date = ?", today).Order("id asc").Find(&metals)

	// If empty, get latest available date
	if len(metals) == 0 {
		var last model.PreciousMetal
		db.DB.Last(&last)
		if last.Date != "" {
			db.DB.Preload("Currency").Where("date = ?", last.Date).Order("id asc").Find(&metals)
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": metals})
}

// GetClientConfig returns client configuration
func (h *Handler) GetClientConfig(c *gin.Context) {
	config.Reload() // Hot Reload
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": config.GetClient(),
	})
}
