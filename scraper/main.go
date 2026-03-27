package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func main() {

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://www.nseindia.com", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Symbol,BookValue,NetProfit,CostOfEquity")
	fmt.Println("INFY,85000,24000,0.13")
	fmt.Println("TCS,102000,42000,0.12")
}
