package api

import (
	"hui_lv_server/internal/model"
	"hui_lv_server/pkg/db"
	"hui_lv_server/pkg/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwtSecret = []byte("hui_lv_secret_key_123") // TODO: Move to config

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

// AdminLogin handles admin authentication
func (h *Handler) AdminLogin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "Invalid request"})
		return
	}

	var admin model.Admin
	if err := db.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "Invalid credentials"})
		return
	}

	// Verify Password
	if utils.HashPassword(req.Password, admin.Salt) != admin.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "Invalid credentials"})
		return
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"username": admin.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": TokenResponse{Token: tokenString}})
}

// AuthMiddleware validates JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "No token provided"})
			return
		}

		// Remove Bearer prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "Invalid token"})
			return
		}

		c.Next()
	}
}

// CRUD Handlers

// GetAdmins
func (h *Handler) GetAdmins(c *gin.Context) {
	username := c.Query("username")
	var admins []model.Admin
	query := db.DB
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	query.Find(&admins)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": admins})
}

// CreateAdmin
func (h *Handler) CreateAdmin(c *gin.Context) {
	var admin model.Admin
	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	// Hash password
	salt := utils.GenerateSalt()
	admin.Salt = salt
	admin.Password = utils.HashPassword(admin.Password, salt)

	if err := db.DB.Create(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": admin})
}

// UpdateAdmin
func (h *Handler) UpdateAdmin(c *gin.Context) {
	var admin model.Admin
	if err := db.DB.First(&admin, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "Admin not found"})
		return
	}
	var req model.Admin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	admin.Username = req.Username
	if req.Password != "" {
		salt := utils.GenerateSalt()
		admin.Salt = salt
		admin.Password = utils.HashPassword(req.Password, salt)
	}
	db.DB.Save(&admin)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": admin})
}

// DeleteAdmin
func (h *Handler) DeleteAdmin(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "Invalid ID"})
		return
	}
	if err := db.DB.Delete(&model.Admin{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Deleted"})
}

// --- Bank Management ---

func (h *Handler) GetBanks(c *gin.Context) {
	keyword := c.Query("keyword")
	var banks []model.Bank
	query := db.DB
	if keyword != "" {
		query = query.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Find(&banks)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": banks})
}

func (h *Handler) CreateBank(c *gin.Context) {
	var bank model.Bank
	if err := c.ShouldBindJSON(&bank); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	db.DB.Create(&bank)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": bank})
}

func (h *Handler) UpdateBank(c *gin.Context) {
	var bank model.Bank
	if err := db.DB.First(&bank, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "Not found"})
		return
	}
	var req model.Bank
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	bank.Code = req.Code
	bank.Name = req.Name
	db.DB.Save(&bank)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": bank})
}

func (h *Handler) DeleteBank(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "Invalid ID"})
		return
	}
	db.DB.Delete(&model.Bank{}, id)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Deleted"})
}

// --- Currency Management ---

func (h *Handler) GetCurrencies(c *gin.Context) {
	keyword := c.Query("keyword")
	var currencies []model.Currency
	query := db.DB.Order("sort asc")
	if keyword != "" {
		query = query.Where("code LIKE ? OR name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	query.Find(&currencies)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": currencies})
}

func (h *Handler) CreateCurrency(c *gin.Context) {
	var cur model.Currency
	if err := c.ShouldBindJSON(&cur); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	db.DB.Create(&cur)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": cur})
}

func (h *Handler) UpdateCurrency(c *gin.Context) {
	var cur model.Currency
	if err := db.DB.First(&cur, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "Not found"})
		return
	}
	var req model.Currency
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	cur.Code = req.Code
	cur.Name = req.Name
	cur.Sort = req.Sort
	db.DB.Save(&cur)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": cur})
}

func (h *Handler) DeleteCurrency(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "Invalid ID"})
		return
	}
	db.DB.Delete(&model.Currency{}, id)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Deleted"})
}

// --- Bank Rate Management (Admin View) ---

func (h *Handler) GetAllBankRates(c *gin.Context) {
    // Pagination
    page := 1
    pageSize := 20
	// Filters
	bankID := c.Query("bank_id")
	currencyID := c.Query("currency_id")
	date := c.Query("date")
    
	var rates []model.BankRate
	query := db.DB.Preload("Bank").Preload("Currency").Order("id desc")

	if bankID != "" {
		query = query.Where("bank_id = ?", bankID)
	}
	if currencyID != "" {
		query = query.Where("currency_id = ?", currencyID)
	}
	if date != "" {
		query = query.Where("date = ?", date)
	}

	query.Limit(pageSize).Offset((page - 1) * pageSize).Find(&rates)
        
    var total int64
	// Allow counting with filters to support pagination in future properly
	// Re-using same filters for count
	countQuery := db.DB.Model(&model.BankRate{})
	if bankID != "" {
		countQuery = countQuery.Where("bank_id = ?", bankID)
	}
	if currencyID != "" {
		countQuery = countQuery.Where("currency_id = ?", currencyID)
	}
	if date != "" {
		countQuery = countQuery.Where("date = ?", date)
	}
    countQuery.Count(&total)

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": rates, "total": total})
}

func (h *Handler) GetDashboardStats(c *gin.Context) {
	var adminCount int64
	var bankCount int64
	var currencyCount int64
	var rateCount int64

	db.DB.Model(&model.Admin{}).Count(&adminCount)
	db.DB.Model(&model.Bank{}).Count(&bankCount)
	db.DB.Model(&model.Currency{}).Count(&currencyCount)
	db.DB.Model(&model.BankRate{}).Count(&rateCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 1, 
		"data": gin.H{
			"admin_count": adminCount,
			"bank_count": bankCount,
			"currency_count": currencyCount,
			"rate_count": rateCount,
		},
	})
}

func (h *Handler) DeleteBankRate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "Invalid ID"})
		return
	}
	db.DB.Delete(&model.BankRate{}, id)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Deleted"})
}

// --- Exchange Rate Management (Read-only/Delete) ---

func (h *Handler) GetExchangeRates(c *gin.Context) {
	// Pagination
	page := 1
	pageSize := 20
	// Filters
	currencyID := c.Query("currency_id")
	date := c.Query("date")

	var rates []model.ExchangeRate
	query := db.DB.Preload("Currency").Order("id desc")

	if currencyID != "" {
		query = query.Where("currency_id = ?", currencyID)
	}
	if date != "" {
		query = query.Where("date = ?", date)
	}

	query.Limit(pageSize).Offset((page - 1) * pageSize).Find(&rates)

	var total int64
	countQuery := db.DB.Model(&model.ExchangeRate{})
	if currencyID != "" {
		countQuery = countQuery.Where("currency_id = ?", currencyID)
	}
	if date != "" {
		countQuery = countQuery.Where("date = ?", date)
	}
	countQuery.Count(&total)

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": rates, "total": total})
}

// --- Bank Deposit Management ---

func (h *Handler) GetDepositsAdmin(c *gin.Context) {
	// Pagination
	page := 1
	pageSize := 20
	// Filters
	bankID := c.Query("bank_id")
	currencyID := c.Query("currency_id")

	var deposits []model.BankDeposit
	query := db.DB.Preload("Bank").Preload("Currency").Order("id desc")

	if bankID != "" {
		query = query.Where("bank_id = ?", bankID)
	}
	if currencyID != "" {
		query = query.Where("currency_id = ?", currencyID)
	}

	query.Limit(pageSize).Offset((page - 1) * pageSize).Find(&deposits)

	var total int64
	countQuery := db.DB.Model(&model.BankDeposit{})
	if bankID != "" {
		countQuery = countQuery.Where("bank_id = ?", bankID)
	}
	if currencyID != "" {
		countQuery = countQuery.Where("currency_id = ?", currencyID)
	}
	countQuery.Count(&total)

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": deposits, "total": total})
}

func (h *Handler) CreateDeposit(c *gin.Context) {
	var deposit model.BankDeposit
	if err := c.ShouldBindJSON(&deposit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	db.DB.Create(&deposit)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": deposit})
}

func (h *Handler) UpdateDeposit(c *gin.Context) {
	var deposit model.BankDeposit
	if err := db.DB.First(&deposit, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "Not found"})
		return
	}
	var req model.BankDeposit
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": err.Error()})
		return
	}
	deposit.BankID = req.BankID
	deposit.CurrencyID = req.CurrencyID
	deposit.Name = req.Name
	deposit.Time = req.Time
	deposit.Rate = req.Rate
	deposit.Tips = req.Tips
	
	db.DB.Save(&deposit)
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": deposit})
}

func (h *Handler) DeleteDeposit(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "Invalid ID"})
		return
	}
	db.DB.Delete(&model.BankDeposit{}, id)
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Deleted"})
}
