package main

import (
	"hui_lv_server/config"
	"hui_lv_server/internal/api"
	"hui_lv_server/internal/service"
	"hui_lv_server/pkg/db"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"time"
)

func main() {
	// Set Timezone to Asia/Shanghai
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		log.Printf("Warning: Could not load Asia/Shanghai timezone: %v", err)
	} else {
		time.Local = loc
	}

	// 0. Init Config
	config.Init()

	// 1. Init DB
	db.Init()
	service.Seed()

	// 2. Start Scheduler
	c := cron.New()
	syncService := &service.SyncService{}

	// Run sync on startup (async)
	go func() {
		log.Println("Performing initial sync...")
		syncService.SyncRealTimeRates()
		syncService.SyncBankRatesFromKylc()
		syncService.SyncPreciousMetals()
		//syncService.SyncAll()
	}()

	// Schedule sync every 4 hours
	//_, err := c.AddFunc("0 */4 * * *", func() {
	//	syncService.SyncAll()
	//})
	//if err != nil {
	//	log.Fatal("Error adding cron job")
	//}

	// Schedule Bank Rates sync every 20 minutes
	_, err = c.AddFunc("*/20 * * * *", func() {
		syncService.SyncBankRatesFromKylc()
	})
	if err != nil {
		log.Printf("Error adding bank rates sync cron job: %v", err)
	}

	c.Start()

	// Schedule sync every 10 minutes for Real-time Rates
	_, err = c.AddFunc("*/10 * * * *", func() {
		syncService.SyncRealTimeRates()
		syncService.SyncPreciousMetals()
	})
	if err != nil {
		log.Printf("Error adding real-time sync cron job: %v", err)
	}

	// 3. Setup Router
	r := gin.Default()

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	handler := api.NewHandler()
	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/rates/latest", handler.GetRates)
		apiGroup.GET("/banks", handler.GetBankRates)
		apiGroup.GET("/history", handler.GetHistory)
		apiGroup.POST("/subscribe", handler.Subscribe)
		apiGroup.GET("/subscriptions", handler.GetSubscriptions)
		apiGroup.GET("/deposits", handler.GetDeposits)
		apiGroup.GET("/metals", handler.GetPreciousMetals)
		apiGroup.GET("/config", handler.GetClientConfig)
		apiGroup.GET("/currencies", handler.GetCurrencies)
	}

	// Admin Routes
	adminGroup := r.Group("/api/admin")
	{
		adminGroup.POST("/login", handler.AdminLogin)
	}
	
	privateAdminGroup := r.Group("/api/admin")
	privateAdminGroup.Use(api.AuthMiddleware())
	{
		// Dashboard
		privateAdminGroup.GET("/stats", handler.GetDashboardStats)

		// Admin Management
		privateAdminGroup.GET("/admins", handler.GetAdmins)
		privateAdminGroup.POST("/admins", handler.CreateAdmin)
		privateAdminGroup.PUT("/admins/:id", handler.UpdateAdmin)
		privateAdminGroup.DELETE("/admins/:id", handler.DeleteAdmin)

		// Bank Management
		privateAdminGroup.GET("/banks", handler.GetBanks)
		privateAdminGroup.POST("/banks", handler.CreateBank)
		privateAdminGroup.PUT("/banks/:id", handler.UpdateBank)
		privateAdminGroup.DELETE("/banks/:id", handler.DeleteBank)

		// Currency Management
		privateAdminGroup.GET("/currencies", handler.GetCurrencies)
		privateAdminGroup.POST("/currencies", handler.CreateCurrency)
		privateAdminGroup.PUT("/currencies/:id", handler.UpdateCurrency)
		privateAdminGroup.DELETE("/currencies/:id", handler.DeleteCurrency)

		// Bank Rate Management (Admin View)
		privateAdminGroup.GET("/bank_rates", handler.GetAllBankRates)
		privateAdminGroup.DELETE("/bank_rates/:id", handler.DeleteBankRate)

		// Real-time Rates Management
		privateAdminGroup.GET("/exchange_rates", handler.GetExchangeRates)

		// Deposit Management
		privateAdminGroup.GET("/deposits", handler.GetDepositsAdmin)
		privateAdminGroup.POST("/deposits", handler.CreateDeposit)
		privateAdminGroup.PUT("/deposits/:id", handler.UpdateDeposit)
		privateAdminGroup.DELETE("/deposits/:id", handler.DeleteDeposit)
	}

	// 4. Run Server
	log.Printf("Server starting on %s", config.AppConfig.Server.Port)
	r.Run(config.AppConfig.Server.Port)
}
