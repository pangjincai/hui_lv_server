package consts

// CurrencyNames maps ISO currency codes to Chinese names
var CurrencyNames = map[string]string{
	"USD": "美元",
	"EUR": "欧元",
	"HKD": "港元",
	"JPY": "日元",
	"GBP": "英镑",
	"AUD": "澳元",
	"CAD": "加元",
	"SGD": "新加坡元",
	"CHF": "瑞士法郎",
	"RUB": "俄罗斯卢布",
	"KRW": "韩元",
	"THB": "泰铢",
	"NZD": "新西兰元",
	"MOP": "澳门元",
	"TWD": "新台币",
	"PHP": "菲律宾比索",
	"MYR": "马来西亚林吉特",
	"DKK": "丹麦克朗",
	"SEK": "瑞典克朗",
	"NOK": "挪威克朗",
	"TRY": "土耳其里拉",
	"BRL": "巴西雷亚尔",
	"INR": "印度卢比",
	"IDR": "印度尼西亚盾",
	"ILS": "以色列新谢克尔",
	"ZAR": "南非兰特",
	"SAR": "沙特里亚尔",
	"AED": "阿联酋迪拉姆",
	"HUF": "匈牙利福林",
	"MXN": "墨西哥比索",
	"PLN": "波兰兹罗提",
	"BND": "文莱元",
	"KZT": "坚戈",
	"KHR": "柬埔寨瑞尔",
}

// BankUrlCodes maps internal bank codes to 5waihui.com URL slugs
var BankUrlCodes = map[string]string{
	"ICBC":     "icbc",
	"BOC":      "boc",
	"ABCHINA":  "abc",
	"BANKCOMM": "bcm",
	"CCB":      "ccb",
	"CMBCHINA": "cmb",
	"CEBBANK":  "ceb",
	"SPDB":     "spdb",
	"CIB":      "cib",
	"ECITIC":   "citic",
	"PSBC":     "psbc",
	"CMBC":     "cmbc",
}

// CurrencyNameToCode maps Chinese currency names to ISO codes
var CurrencyNameToCode = map[string]string{}

func init() {
	for code, name := range CurrencyNames {
		CurrencyNameToCode[name] = code
	}
	// Add common aliases
	CurrencyNameToCode["阿联酋币"] = "AED"
	CurrencyNameToCode["巴西币"] = "BRL"
	CurrencyNameToCode["加币"] = "CAD" // Alias for 加元
	CurrencyNameToCode["加拿大币"] = "CAD"
	CurrencyNameToCode["瑞士法郎"] = "CHF"
	CurrencyNameToCode["丹麦克朗"] = "DKK"
	CurrencyNameToCode["匈牙利币"] = "HUF" // Assuming HUF, need to check if in MajorCurrencies
	CurrencyNameToCode["印尼盾"] = "IDR"
	CurrencyNameToCode["以色列币"] = "ILS"
	CurrencyNameToCode["印度卢比"] = "INR"
	CurrencyNameToCode["柬埔寨币"] = "KHR" // Not in MajorCurrencies?
	CurrencyNameToCode["瑞郎"] = "CHF"
	CurrencyNameToCode["新西兰币"] = "NZD"
	CurrencyNameToCode["文莱元"] = "BND"
	CurrencyNameToCode["捷克克朗"] = "CZK"
	CurrencyNameToCode["坚戈"] = "KZT"
	CurrencyNameToCode["墨西哥比索"] = "MXN"
	CurrencyNameToCode["波兰兹罗提"] = "PLN"
	CurrencyNameToCode["马币"] = "MYR"
	CurrencyNameToCode["林吉特"] = "MYR"
	CurrencyNameToCode["马来西亚林吉特"] = "MYR"
	CurrencyNameToCode["菲律宾币"] = "PHP"
	CurrencyNameToCode["沙特币"] = "SAR"
	CurrencyNameToCode["土耳其币"] = "TRY"
	CurrencyNameToCode["美国美元"] = "USD"
	CurrencyNameToCode["港币"] = "HKD" // Alias for 港元
	CurrencyNameToCode["台币"] = "TWD" // Alias for 新台币
	CurrencyNameToCode["韩国元"] = "KRW"
	CurrencyNameToCode["韩币"] = "KRW"
	CurrencyNameToCode["澳门币"] = "MOP"
	CurrencyNameToCode["卢布"] = "RUB"
	CurrencyNameToCode["新加坡币"] = "SGD"
	CurrencyNameToCode["巴西里亚尔"] = "BRL"
}
