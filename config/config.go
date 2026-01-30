package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
)

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

type TanshuConfig struct {
	ApiKey string `json:"api_key"`
	ApiUrl string `json:"api_url"`
}

type ClientConfig struct {
	ShowAd   bool   `json:"show_ad"`
	AdUnitID string `json:"ad_unit_id"`
}

type Config struct {
	Database DatabaseConfig `json:"database"`
	Server   ServerConfig   `json:"server"`
	Tanshu   TanshuConfig   `json:"tanshu"`
	Client   ClientConfig   `json:"client"`
}

var (
	AppConfig Config
	configMu  sync.RWMutex
)

// Init loads configuration from config.json
func Init() {
	Reload()
}

// Reload reads config.json and updates AppConfig
func Reload() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal("Error reading config.json: ", err)
	}

	configMu.Lock()
	defer configMu.Unlock()
	err = json.Unmarshal(data, &AppConfig)
	if err != nil {
		log.Printf("Error parsing config.json: %v", err)
		return
	}
}

// GetClient returns a copy of ClientConfig safely
func GetClient() ClientConfig {
	configMu.RLock()
	defer configMu.RUnlock()
	return AppConfig.Client
}

// GetServer returns a copy of ServerConfig safely
func GetServer() ServerConfig {
	configMu.RLock()
	defer configMu.RUnlock()
	return AppConfig.Server
}

// Backward compatibility constants (optional, but better to update usages)
// or just remove them to force update. I will remove them to ensure I update all references.
