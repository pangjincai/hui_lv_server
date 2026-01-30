package model

type Bank struct {
	ID   uint   `gorm:"primaryKey" json:"ID"`
	Code string `gorm:"type:varchar(20);uniqueIndex" json:"Code"` // e.g., ICBC
	Name string `json:"Name"` // e.g., 中国工商银行
}
