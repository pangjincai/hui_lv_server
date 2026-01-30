package model

type Currency struct {
	ID   uint   `gorm:"primaryKey" json:"ID"`
	Code string `gorm:"type:varchar(20);uniqueIndex" json:"Code"` // e.g., USD
	Name string `json:"Name"` // e.g., 美元
	Sort int    `gorm:"default:0" json:"Sort"` // Sort order, smaller is higher priority
}
