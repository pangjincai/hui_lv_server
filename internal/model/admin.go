package model

import (
	"gorm.io/gorm"
)

// Admin represents an administrator user
type Admin struct {
	ID        uint      `gorm:"primaryKey" json:"ID"`
	CreatedAt LocalTime `json:"CreatedAt"`
	UpdatedAt LocalTime `json:"UpdatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Username string `gorm:"type:varchar(100);uniqueIndex;not null" json:"Username"`
	Password string `gorm:"not null" json:"-"` // Hashed password
	Salt     string `gorm:"not null" json:"Salt"` // Random salt
}
