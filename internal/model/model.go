package model

import (
	"gorm.io/gorm"
)

// ExchangeRate stores basic currency exchange rates
type ExchangeRate struct {
	ID        uint      `gorm:"primaryKey" json:"ID"`
	CreatedAt LocalTime `json:"CreatedAt"`
	UpdatedAt LocalTime `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Date       string `gorm:"index" json:"Date"` // YYYY-MM-DD
	CurrencyID uint   `gorm:"index" json:"CurrencyID"`
	Currency   Currency `json:"Currency"`
	Base       string  `gorm:"default:'CNY'" json:"Base"` // Base Currency
	Rate       float64 `json:"Rate"` // Exchange rate
}

// BankRate stores bank-specific exchange rates
type BankRate struct {
	ID        uint      `gorm:"primaryKey" json:"ID"`
	CreatedAt LocalTime `json:"CreatedAt"`
	UpdatedAt LocalTime `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Date       string `gorm:"index" json:"Date"`
	BankID     uint   `gorm:"index" json:"BankID"`
	Bank       Bank   `json:"Bank"`
	CurrencyID uint   `gorm:"index" json:"CurrencyID"`
	Currency   Currency `json:"Currency"`
	HuiIn      float64 `json:"HuiIn"` // 现汇买入
	ChaoIn     float64 `json:"ChaoIn"` // 现钞买入
	HuiOut     float64 `json:"HuiOut"` // 现汇卖出
	ChaoOut    float64 `json:"ChaoOut"` // 现钞卖出
	Zhesuan    float64 `json:"Zhesuan"` // 中间价
}

// UserSubscription stores user price alerts
type UserSubscription struct {
	ID        uint      `gorm:"primaryKey" json:"ID"`
	CreatedAt LocalTime `json:"CreatedAt"`
	UpdatedAt LocalTime `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	OpenID   string `gorm:"index"`
	Currency string // Keep as string for now if it's just a code, or link to ID? User request said "link API interfaces", maybe this is less critical but consistency is good. Leaving as string for now as it's user input driven, or maybe switch to ID later. Plan said "Update API handlers". Let's stick to core tables first. Actually, for subscription, keeping as string code might be easier for matching unless we join every time. The prompt focused on "existing tables via ID... banks and currencies tables". "UserSubscription" wasn't explicitly mentioned but "redundant fields" implies cleanup.
	// Let's leave UserSubscription alone for a second or update it if it uses the same "Currency" concept. The prompt said "original table... linked by ID".
	// Let's assume UserSubscription is also a target.
	// However, `UserSubscription` uses `Currency` string. Changing it to ID might break things if not careful.
	// But `ExchangeRate` and `BankRate` are the main data tables.
	// Let's stick to the prompt's "banks" and "currencies" tables and "original table... linked by ID". The "original table" likely refers to the main data tables.
	// I will update ExchangeRate and BankRate primarily.
	Threshold  float64
	NotifyType string // "Daily" or "Threshold"
}

// BankDeposit stores foreign currency deposit products
type BankDeposit struct {
	ID        uint      `gorm:"primaryKey" json:"ID"`
	CreatedAt LocalTime `json:"CreatedAt"`
	UpdatedAt LocalTime `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	BankID     uint `gorm:"index" json:"BankID"`
	Bank       Bank `json:"Bank"`
	CurrencyID uint `gorm:"index" json:"CurrencyID"`
	Currency   Currency `json:"Currency"`
	Name       string `gorm:"column:title" json:"Name"` // Product Name
	Time       string `json:"Time"` // Duration
	Rate       string `json:"Rate"` // Interest Rate description
	Tips       string `json:"Tips"` // Special notes
}

// PreciousMetal stores gold/silver prices
type PreciousMetal struct {
	ID           uint      `gorm:"primaryKey" json:"ID"`
	CreatedAt    LocalTime `json:"CreatedAt"`
	UpdatedAt    LocalTime `json:"UpdatedAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Date         string  `gorm:"index" json:"Date"` // YYYY-MM-DD
	TradingVenue string  `gorm:"column:trading_venue;index" json:"TradingVenue"` // previously Exchange
	Symbol       string  `json:"Symbol"`
	Title        string  `gorm:"column:title" json:"Title"` // previously Name
	Type         string  `gorm:"column:type" json:"Type"` // 黄金/白银
	Price        float64 `json:"Price"`
	MinPrice     float64 `gorm:"column:min_price" json:"MinPrice"` // Low
	MaxPrice     float64 `gorm:"column:max_price" json:"MaxPrice"` // High
	KpPrice      float64 `gorm:"column:kp_price" json:"KpPrice"`   // Open
	PrevClose    float64 `json:"PrevClose"`
	Unit         string  `json:"Unit"`
	CurrencyID   uint    `gorm:"index" json:"CurrencyID"`
	Currency     Currency `json:"Currency"`
}
