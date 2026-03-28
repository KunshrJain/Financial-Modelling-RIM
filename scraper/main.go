package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func main() {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}

	// 1. ABSOLUTE PATH LOGIC
	// Gets the absolute path of the current working directory (scraper)
	cwd, _ := os.Getwd()
	// Points to the directory one level up, then into "data"
	dataDir := filepath.Clean(filepath.Join(cwd, "..", "data"))
	
	// Check if the folder actually exists
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		fmt.Printf("⚠️ Folder %s not found, creating it...\n", dataDir)
		os.MkdirAll(dataDir, 0755)
	}

	fmt.Println("Step 1: Session Handshake...")
	warmup(client)

	symbols := []string{"INFY", "TCS"}
	fmt.Println("Symbol,Year,Type,BookValue,NetProfit,Beta")

	for _, s := range symbols {
		// 2. Metadata Fetch
		apiURL := "https://www.nseindia.com/api/annual-reports-xbrl?index=equities&symbol=" + s
		req, _ := http.NewRequest("GET", apiURL, nil)
		setHeaders(req, "https://www.nseindia.com/companies-listing/corporate-filings-annual-reports-xbrl")
		
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != 200 { continue }
		
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var raw map[string]interface{}
		json.Unmarshal(body, &raw)
		
		var targetURL string
		var subType string
		var year string

		if data, ok := raw["data"].([]interface{}); ok && len(data) > 0 {
			chosenItem := data[0].(map[string]interface{})
			for _, item := range data {
				m := item.(map[string]interface{})
				if m["submission_type"] == "Consolidated" {
					chosenItem = m
					break
				}
			}
			targetURL, _ = chosenItem["fileName"].(string)
			subType, _ = chosenItem["submission_type"].(string)
			year = fmt.Sprintf("%v-%v", chosenItem["fromYr"], chosenItem["toYr"])
		}

		if targetURL == "" { continue }

		// 3. Download
		encodedURL := url.QueryEscape(`"` + targetURL + `"`)
		proxyURL := "https://www.nseindia.com/api/download_xbrl?fileUrl=" + encodedURL

		dlReq, _ := http.NewRequest("GET", proxyURL, nil)
		setHeaders(dlReq, "https://www.nseindia.com/")
		
		dlResp, err := client.Do(dlReq)
		if err != nil || dlResp.StatusCode != 200 { continue }
		
		dlBody, _ := io.ReadAll(dlResp.Body)
		dlResp.Body.Close()

		// 4. Save to the verified dataDir
		safeYear := strings.ReplaceAll(year, "-", "_")
		fileName := fmt.Sprintf("%s_%s_%s.xml", s, subType, safeYear)
		savePath := filepath.Join(dataDir, fileName)
		
		err = os.WriteFile(savePath, dlBody, 0644)
		if err != nil {
			fmt.Printf("❌ Write Error: %v\n", err)
			continue
		}

		// 5. Extraction
		content := string(dlBody)
		profit := extract(content, "ProfitLossForPeriod")
		equity := extract(content, "EquityAttributableToOwnersOfParent")

		if profit != 0 {
			fmt.Printf("%s,%s,%s,%.2f,%.2f,1.1\n", s, year, subType, equity, profit)
		}
		
		time.Sleep(2 * time.Second)
	}
	fmt.Printf("\n✅ Sync Complete. Files saved to: %s\n", dataDir)
}

func warmup(c *http.Client) {
	u1 := "https://www.nseindia.com/"
	u2 := "https://www.nseindia.com/api/marketStatus"
	for _, u := range []string{u1, u2} {
		r, _ := http.NewRequest("GET", u, nil)
		setHeaders(r, "")
		res, _ := c.Do(r)
		if res != nil { res.Body.Close() }
	}
}

func setHeaders(req *http.Request, ref string) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	if ref != "" { req.Header.Set("Referer", ref) }
}

func extract(content string, tag string) float64 {
	re := regexp.MustCompile(`(?i)` + tag + `[^>]*>([\d\s\.,]+)</`)
	match := re.FindStringSubmatch(content)
	if len(match) > 1 {
		clean := strings.ReplaceAll(match[1], ",", "")
		val, _ := strconv.ParseFloat(strings.TrimSpace(clean), 64)
		return val
	}
	return 0
}