package service

import (
	"bytes"
	"hui_lv_server/internal/consts"
	"hui_lv_server/internal/model"
	"hui_lv_server/pkg/db"
	"hui_lv_server/pkg/request"
	"log"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// SyncPreciousMetals scrapes gold/silver prices from huilvbiao.com
func (s *SyncService) SyncPreciousMetals() {
	log.Println("Starting SyncPreciousMetals...")
	url := "https://www.huilvbiao.com/api/gold_indexApi"
	body, err := request.GetRaw(url)
	if err != nil {
		log.Printf("Failed to fetch precious metals: %v", err)
		return
	}
	content := string(body)
	today := time.Now().Format("2006-01-02")

	// --- Gold ---
	// 1. Shanghai Gold Exchange (AUTD)
	// Index: 0=Price, 4=High, 5=Low, 7=PrevClose (Yesterday Settlement), 8=Open
	// var hq_str_gds_AUTD="1247.73,0,1247.12,1247.41,1255.00,1177.25,14:50:50,1184.04,1180.44,102610,3.00,5.00,2026-01-29,黄金延期";
	s.parseAndSaveMetal(content, "hq_str_gds_AUTD", "中国上海黄金交易所", "AUTD", "国内黄金价格", "克", "CNY", "黄金", 0, 4, 5, 8, 7, today)

	// 2. COMEX Gold (GC)
	// var hq_str_hf_GC="5599.955,,5596.900,5597.400,5626.800,5449.400,14:50:54,5340.200,5449.900,0,1,1,2026-01-29,纽约黄金,0";
	s.parseAndSaveMetal(content, "hq_str_hf_GC", "美国纽约商品交易所", "GC", "纽约期货国际金价", "盎司", "USD", "黄金", 0, 4, 5, 8, 7, today)

	// 3. London Gold (XAU)
	// var hq_str_hf_XAU="5567.76,5414.840,5567.76,5568.06,5596.33,5422.62,14:50:00,5414.84,5422.62,0,0,0,2026-01-29,伦敦金（现货黄金）";
	s.parseAndSaveMetal(content, "hq_str_hf_XAU", "英国伦敦黄金交易市场", "XAU", "伦敦现货黄金价格", "盎司", "USD", "黄金", 0, 4, 5, 8, 7, today)

	// --- Silver ---
	urlSilver := "https://www.huilvbiao.com/api/silver_indexApi"
	bodySilver, err := request.GetRaw(urlSilver)
	if err != nil {
		log.Printf("Failed to fetch silver prices: %v", err)
	} else {
		contentSilver := string(bodySilver)
		// 1. Shanghai Gold Exchange (AGTD)
		// var hq_str_gds_AGTD="29925.00,0,29943.00,30099.00,30280.00,28750.00,15:00:46,29310.00,29018.00,672860,10.00,20.00,2026-01-29,白银延期";
		s.parseAndSaveMetal(contentSilver, "hq_str_gds_AGTD", "中国上海黄金交易所", "AGTD", "国内白银价格", "千克", "CNY", "白银", 0, 4, 5, 8, 7, today)

		// 2. COMEX Silver (SI)
		// var hq_str_hf_SI="119.213,,119.075,119.115,120.565,115.500,15:01:14,113.534,116.625,0,1,2,2026-01-29,纽约白银,0";
		s.parseAndSaveMetal(contentSilver, "hq_str_hf_SI", "美国纽约商品交易所", "SI", "纽约期货国际银价", "盎司", "USD", "白银", 0, 4, 5, 8, 7, today)

		// 3. London Silver (XAG)
		// var hq_str_hf_XAG="118.80,116.691,118.80,119.14,120.28,115.29,15:01:00,116.69,116.59,0,0,0,2026-01-29,伦敦银（现货白银）";
		s.parseAndSaveMetal(contentSilver, "hq_str_hf_XAG", "英国伦敦黄金交易市场", "XAG", "伦敦现货白银价格", "盎司", "USD", "白银", 0, 4, 5, 8, 7, today)
	}

	log.Println("SyncPreciousMetals Finished.")
}

func (s *SyncService) parseAndSaveMetal(content, varName, exchange, symbol, title, unit, currencyCode, typeName string, idxPrice, idxHigh, idxLow, idxOpen, idxPrevClose int, date string) {
	re := regexp.MustCompile(varName + `="([^"]+)"`)
	match := re.FindStringSubmatch(content)
	if len(match) < 2 {
		log.Printf("Variable %s not found", varName)
		return
	}
	parts := strings.Split(match[1], ",")
	
	// Need at least up to max index
	maxIdx := idxPrice
	for _, idx := range []int{idxHigh, idxLow, idxOpen, idxPrevClose} {
		if idx > maxIdx { maxIdx = idx }
	}

	if len(parts) <= maxIdx {
		log.Printf("Variable %s format error: insufficient parts, need %d, got %d", varName, maxIdx, len(parts))
		return
	}

	price := parseFloat(parts[idxPrice])
	high := parseFloat(parts[idxHigh])
	low := parseFloat(parts[idxLow])
	open := parseFloat(parts[idxOpen])
	prevClose := parseFloat(parts[idxPrevClose])

	if price == 0 {
		return
	}

	// Get CurrencyID
	// Special handling for CNY (not in standard scraped currencies list usually)
	var currencyID uint
	if currencyCode == "CNY" {
		// Try to find CNY or create it if missing
		var c model.Currency
		if err := db.DB.Where("code = ?", "CNY").First(&c).Error; err != nil {
			// Create CNY
			c = model.Currency{Code: "CNY", Name: "人民币", Sort: 0}
			db.DB.Create(&c)
		}
		currencyID = c.ID
	} else {
		currencyID = s.getCurrencyID(currencyCode)
	}

	record := model.PreciousMetal{
		Date:         date,
		TradingVenue: exchange,
		Symbol:       symbol,
		Title:        title,
		Type:         typeName,
		Price:        price,
		MaxPrice:     high,
		MinPrice:     low,
		KpPrice:      open,
		PrevClose:    prevClose,
		Unit:         unit,
		CurrencyID:   currencyID,
	}

	// Update or Create
	var count int64
	db.DB.Model(&model.PreciousMetal{}).
		Where("date = ? AND trading_venue = ? AND symbol = ?", date, exchange, symbol).
		Count(&count)

	if count == 0 {
		db.DB.Create(&record)
	} else {
		db.DB.Model(&model.PreciousMetal{}).
			Where("date = ? AND trading_venue = ? AND symbol = ?", date, exchange, symbol).
			Updates(&record)
	}
}

// SyncBankRatesFromKylc scrapes bank rates from kylc.com
func (s *SyncService) SyncBankRatesFromKylc() {
	log.Println("Starting SyncBankRatesFromKylc...")
	today := time.Now().Format("2006-01-02")
	currentYear := time.Now().Year()

	// 1. Fetch Bank List
	bankMap, err := s.fetchBankListFromKylc()
	if err != nil {
		log.Printf("Failed to fetch bank list: %v", err)
		return
	}

	for bankName, bankUrl := range bankMap {
		// 2. Get Bank ID by Name
		bankID := s.getBankIDByName(bankName)
		if bankID == 0 {
			// Auto-create bank if not found
			// Extract code from URL: .../b-{code}.html
			reCode := regexp.MustCompile(`b-([a-zA-Z0-9]+)\.html$`)
			matches := reCode.FindStringSubmatch(bankUrl)
			if len(matches) > 1 {
				newCode := strings.ToUpper(matches[1])
				log.Printf("Auto-creating bank: %s (%s)", bankName, newCode)
				
				newBank := model.Bank{
					Name: bankName,
					Code: newCode,
				}
				if err := db.DB.Create(&newBank).Error; err == nil {
					bankID = newBank.ID
				} else {
					log.Printf("Failed to create bank %s: %v", bankName, err)
					continue
				}
			} else {
				// log.Printf("Bank not found in DB and failed to extract code: %s (%s)", bankName, bankUrl)
				continue
			}
		}

		// 3. Fetch Bank Detail Page
		if !strings.HasPrefix(bankUrl, "http") {
			bankUrl = "https://www.kylc.com" + bankUrl
		}
		
		body, err := request.GetRaw(bankUrl)
		if err != nil {
			log.Printf("Failed to fetch %s: %v", bankUrl, err)
			continue
		}

		// kylc is UTF-8, no conversion needed usually.
		html := string(body)

		// 4. Parse Rates
		rates := s.parseKylcBankPage(html, bankName, currentYear)
		if len(rates) == 0 {
			log.Printf("No rates found for %s", bankName)
			continue
		}

		for _, item := range rates {
			// Map currency name to code
			currencyCode := item.CurrencyCode
			if code, ok := consts.CurrencyNameToCode[currencyCode]; ok {
				currencyCode = code
			}

			currencyID := s.getCurrencyID(currencyCode)
			if currencyID == 0 {
				continue
			}

			record := model.BankRate{
				Date:       today,
				BankID:     bankID,
				CurrencyID: currencyID,
				HuiIn:      round6(item.HuiIn / 100),
				ChaoIn:     round6(item.ChaoIn / 100),
				HuiOut:     round6(item.HuiOut / 100),
				ChaoOut:    round6(item.ChaoOut / 100),
				Zhesuan:    round6(item.Zhesuan / 100),
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
	log.Println("SyncBankRatesFromKylc Finished.")
}

func (s *SyncService) fetchBankListFromKylc() (map[string]string, error) {
	url := "https://www.kylc.com/huilv/bank.html"
	body, err := request.GetRaw(url)
	if err != nil {
		return nil, err
	}
	html := string(body)
	
	// Regex to extract bank links: <li class="item bank_item">\s*<a href="(.*?)"[^>]*>(.*?)</a>
	re := regexp.MustCompile(`<li class="item bank_item">\s*<a href="([^"]+)"[^>]*>([^<]+)</a>`)
	matches := re.FindAllStringSubmatch(html, -1)
	
	result := make(map[string]string)
	for _, match := range matches {
		if len(match) >= 3 {
			link := match[1]
			name := strings.TrimSpace(match[2])
			result[name] = link
		}
	}
	return result, nil
}

func (s *SyncService) getBankIDByName(name string) uint {
	// Alias mapping for Kylc names vs DB names
	aliases := map[string]string{
		"工商银行": "中国工商银行",
		"农业银行": "中国农业银行",
		"建设银行": "中国建设银行",
		"邮储银行": "中国邮政储蓄银行",
		"浦发银行": "上海浦东发展银行",
		"光大银行": "中国光大银行",
		"民生银行": "中国民生银行",
	}
	
	searchName := name
	if val, ok := aliases[name]; ok {
		searchName = val
	}

	var b model.Bank
	// Try exact match first
	if err := db.DB.Where("name = ?", searchName).First(&b).Error; err == nil {
		return b.ID
	}
	
	// Try fuzzy match as fallback (e.g. "光大银行" matches "中国光大银行")
	if err := db.DB.Where("name LIKE ?", "%"+name+"%").First(&b).Error; err == nil {
		return b.ID
	}
	
	return 0
}

func (s *SyncService) parseKylcBankPage(html string, bankName string, year int) []ScrapedRate {
	// Normalize
	html = strings.ReplaceAll(html, "\n", "")
	html = strings.ReplaceAll(html, "\r", "")

	// Regex for table rows in tbody
	// Simplified: look for trs that contain rate data
	rowRe := regexp.MustCompile(`<tr>(.*?)</tr>`)
	cellRe := regexp.MustCompile(`<td[^>]*>(.*?)</td>`)
	
	rows := rowRe.FindAllStringSubmatch(html, -1)
	var results []ScrapedRate

	for _, row := range rows {
		cells := cellRe.FindAllStringSubmatch(row[1], -1)
		// We expect roughly 6-8 columns
		// Col 0: Currency Name (e.g. 港币, or <b>港币</b>)
		// Col 1: HuiIn
		// Col 2: ChaoIn
		// Col 3: HuiOut
		// Col 4: ChaoOut
		// Col 5: Time (MM-dd HH:mm)
		
		if len(cells) < 6 {
			continue
		}

		// Parse Currency Name
		rawName := removeTags(cells[0][1])
		currencyName := strings.TrimSpace(rawName)
		currencyName = strings.ReplaceAll(currencyName, "&nbsp;", "")
		
		if currencyName == "" || currencyName == "币种" {
			continue
		}

		rate := ScrapedRate{
			CurrencyCode: currencyName, // We use Name as Code temporarily, mapped later
			HuiIn:        parseFloat(cells[1][1]),
			ChaoIn:       parseFloat(cells[2][1]),
			HuiOut:       parseFloat(cells[3][1]),
			ChaoOut:      parseFloat(cells[4][1]),
		}
		
		// Handle missing ChaoIn/ChaoOut (copy from Hui)
		if rate.ChaoIn == 0 { rate.ChaoIn = rate.HuiIn }
		if rate.ChaoOut == 0 { rate.ChaoOut = rate.HuiOut }
		
		// Zhesuan? Not always present in top columns, maybe check if there are more columns
		// But basic requirements are In/Out.
		// If Middle rate exists?
		// Looking at BOC example: 8 columns?
		// <tr><td>港币</td><td>0.8898</td><td>0.8898</td><td>0.8934</td><td>0.8934</td><td>01-27 20:15</td><td>...</td><td>...</td></tr>
		// So 5 columns of data + time.
		// No explicit Zhesuan in this table?
		// That's fine. We calculate it or leave it 0.
		// Actually, middle rate = (In + Out) / 2? Or just leave 0.
		rate.Zhesuan = (rate.HuiIn + rate.HuiOut) / 2

		results = append(results, rate)
	}
	return results
}

func round6(val float64) float64 {
	return math.Round(val*1000000) / 1000000
}

type ScrapedRate struct {
	CurrencyCode string
	HuiIn        float64
	ChaoIn       float64
	HuiOut       float64
	ChaoOut      float64
	Zhesuan      float64
}

func (s *SyncService) parseBankPage(html string, bankCode string) []ScrapedRate {
	// Debug log
	log.Printf("[%s] Parsing HTML length: %d", bankCode, len(html))

	// Normalize HTML
	html = strings.ReplaceAll(html, "\n", "")
	html = strings.ReplaceAll(html, "\r", "")
	
	// Regex for table rows
	rowRe := regexp.MustCompile(`<tr[^>]*>(.*?)</tr>`)
	// Handle unclosed td/th at the end of row
	cellRe := regexp.MustCompile(`<(?:td|th)[^>]*>(.*?)(?:</(?:td|th)>|$)`)
	hrefRe := regexp.MustCompile(`href=".*?/([a-zA-Z]+)/"`)

	rows := rowRe.FindAllStringSubmatch(html, -1)
	if len(rows) == 0 {
		log.Printf("[%s] No rows found", bankCode)
		return nil
	}
	log.Printf("[%s] Found %d rows", bankCode, len(rows))

	var results []ScrapedRate
	
	// Column Indices
	idxCurrency := -1
	idxHuiIn := -1
	idxChaoIn := -1
	idxHuiOut := -1
	idxChaoOut := -1
	idxZhesuan := -1

	for _, row := range rows {
		cells := cellRe.FindAllStringSubmatch(row[1], -1)
		if len(cells) == 0 {
			continue
		}

		// Extract text from cells
		var cellTexts []string
		var currencyCodeFromUrl string

		for k, cell := range cells {
			// Extract href from the first cell (currency name cell)
			if k == 0 { // Assuming currency is always first? Or we check if it matches header
				matches := hrefRe.FindStringSubmatch(cell[1])
				if len(matches) > 1 {
					currencyCodeFromUrl = strings.ToUpper(matches[1])
				}
			}

			// Remove tags inside cell
			text := removeTags(cell[1])
			text = strings.TrimSpace(text)
			cellTexts = append(cellTexts, text)
		}

		// Header detection
		if idxCurrency == -1 {
			for j, text := range cellTexts {
				// Normalize text for header check (remove spaces)
				normText := strings.ReplaceAll(text, " ", "")
				normText = strings.ReplaceAll(normText, "　", "") // Remove full-width space
				normText = strings.ReplaceAll(normText, "&nbsp;", "")

				// log.Printf("[%s] Header Check Col %d: %s", bankCode, j, normText) // Debug

				if strings.Contains(normText, "币种") || strings.Contains(normText, "货币") {
					idxCurrency = j
				} else if strings.Contains(normText, "现汇买入") || strings.Contains(normText, "汇买价") || strings.Contains(normText, "结汇价") || strings.Contains(normText, "结汇") {
					idxHuiIn = j
				} else if strings.Contains(normText, "现钞买入") || strings.Contains(normText, "钞买价") {
					idxChaoIn = j
				} else if strings.Contains(normText, "现汇卖出") || strings.Contains(normText, "汇卖价") || strings.Contains(normText, "购汇价") || strings.Contains(normText, "购汇") {
					idxHuiOut = j
				} else if strings.Contains(normText, "现钞卖出") || strings.Contains(normText, "钞卖价") {
					idxChaoOut = j
				} else if strings.Contains(normText, "折算价") || strings.Contains(normText, "中间价") || strings.Contains(normText, "基准价") {
					idxZhesuan = j
				}
			}
			if idxCurrency != -1 {
				continue
			} else {
				// Fallback: If we found rate columns but not currency column (e.g. ICBC has empty currency header)
				// Assume column 0 is currency if we found at least HuiIn or ChaoIn
				if (idxHuiIn != -1 || idxChaoIn != -1) && idxCurrency == -1 {
					idxCurrency = 0
					log.Printf("Header fallback for %s: Assumed Currency=0, HuiIn=%d, HuiOut=%d", bankCode, idxHuiIn, idxHuiOut)
					continue
				}

				// Hardcode fallbacks for specific banks if header detection fails completely
				if bankCode == "PSBC" {
					idxCurrency = 0
					idxHuiIn = 1
					idxChaoIn = 2
					idxHuiOut = 3
					idxChaoOut = 4
					log.Printf("Hardcoded fallback for PSBC")
					continue
				}
				if bankCode == "ECITIC" {
					idxCurrency = 0
					idxHuiIn = 1
					idxHuiOut = 2
					// ChaoIn/Out will be filled by fallback logic later
					log.Printf("Hardcoded fallback for ECITIC")
					continue
				}
			}
		}

		// Data row parsing
		if idxCurrency != -1 && len(cellTexts) > idxCurrency {
			name := cellTexts[idxCurrency]
			if name == "" {
				continue
			}

			// Prefer code from URL if available and this is the currency column
			// Note: We extracted currencyCodeFromUrl from the *first* cell. 
			// Usually idxCurrency is 0.
			finalCode := name
			if idxCurrency == 0 && currencyCodeFromUrl != "" {
				finalCode = currencyCodeFromUrl
			}

			rate := ScrapedRate{
				CurrencyCode: finalCode,
			}

			if idxHuiIn != -1 && idxHuiIn < len(cellTexts) {
				rate.HuiIn = parseFloat(cellTexts[idxHuiIn])
			}
			if idxChaoIn != -1 && idxChaoIn < len(cellTexts) {
				rate.ChaoIn = parseFloat(cellTexts[idxChaoIn])
			} else if idxHuiIn != -1 { // Fallback if ChaoIn column missing
				rate.ChaoIn = rate.HuiIn
			}

			if idxHuiOut != -1 && idxHuiOut < len(cellTexts) {
				rate.HuiOut = parseFloat(cellTexts[idxHuiOut])
			}
			if idxChaoOut != -1 && idxChaoOut < len(cellTexts) {
				rate.ChaoOut = parseFloat(cellTexts[idxChaoOut])
			} else if idxHuiOut != -1 { // Fallback if ChaoOut column missing
				rate.ChaoOut = rate.HuiOut
			}

			if idxZhesuan != -1 && idxZhesuan < len(cellTexts) {
				rate.Zhesuan = parseFloat(cellTexts[idxZhesuan])
			}

			// PSBC Special Handling: Rates are per 1 unit for most currencies, but per 100 for JPY (named "100日元")
			// We normalize everything to "per 100 units" to match other banks and SyncService logic
			if bankCode == "PSBC" && !strings.Contains(name, "100") {
				rate.HuiIn *= 100
				rate.ChaoIn *= 100
				rate.HuiOut *= 100
				rate.ChaoOut *= 100
				rate.Zhesuan *= 100
			}

			// log.Printf("[%s] Parsed Rate: %+v", bankCode, rate) // Debug
			if rate.CurrencyCode != "" {
				results = append(results, rate)
			}
		}
	}

	return results
}

func removeTags(input string) string {
	// Simple tag removal
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

func parseFloat(input string) float64 {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, ",", "")
	input = strings.ReplaceAll(input, "&nbsp;", "")
	if input == "-" || input == "" {
		return 0
	}
	val, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0
	}
	return val
}

func GbkToUtf8(s []byte) ([]byte, error) {
	cmd := exec.Command("iconv", "-c", "-f", "GBK", "-t", "UTF-8")
	cmd.Stdin = bytes.NewReader(s)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
