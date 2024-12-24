package main

import (
	"context"
	"encoding/json"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CotacaoAPI struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Cotacao struct {
	ID  int `gorm:"primaryKey"`
	Bid float64
	gorm.Model
}

func main() {
	http.HandleFunc("/cotacao", handler)
	log.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		log.Println("Error creating request:", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Error: context deadline exceeded during HTTP request")
		} else {
			log.Println("Error performing HTTP request:", err)
		}
		http.Error(w, "Failed to fetch data", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		log.Println("Error reading response body:", err)
		return
	}

	var cotacao CotacaoAPI
	err = json.Unmarshal(data, &cotacao)
	if err != nil {
		http.Error(w, "Failed to parse JSON", http.StatusInternalServerError)
		log.Println("Error parsing JSON:", err)
		return
	}

	bidValue, err := strconv.ParseFloat(cotacao.Usdbrl.Bid, 64)
	if err != nil {
		http.Error(w, "Failed to convert bid to float", http.StatusInternalServerError)
		log.Println("Error converting bid to float:", err)
		return
	}

	response := map[string]float64{
		"bid": bidValue,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
		log.Println("Error encoding JSON:", err)
		return
	}

	db := connectDB()
	addCotacao(db, Cotacao{Bid: bidValue})
}

func connectDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	db.AutoMigrate(&Cotacao{})
	return db
}

func addCotacao(db *gorm.DB, cotacao Cotacao) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := db.WithContext(ctx).Create(&cotacao).Error
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Error: context deadline exceeded while saving data to database")
		} else {
			log.Println("Error saving data to database:", err)
		}
	}
}
