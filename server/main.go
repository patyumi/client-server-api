package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type USDBRL struct {
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
}

type APIResponse struct {
	USDBRL USDBRL `json:"USDBRL"`
}

func main() {
	http.HandleFunc("/cotacao", CotacaoDolarHandler)
	http.ListenAndServe(":8080", nil)
}

func CotacaoDolarHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cambio, err := CotacaoDolar(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(*cambio)
}

func CotacaoDolar(ctx context.Context) (*string, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	select {
	case <-ctxTimeout.Done():
		log.Println("[CotacaoDolar] Requisição cancelada pelo servidor: Tempo de execução insuficiente")
	default:
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var result APIResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	err = RegistrarCotacao(ctx, &result.USDBRL)
	if err != nil {
		return nil, err
	}

	return &result.USDBRL.Bid, nil
}

func RegistrarCotacao(ctx context.Context, cotacao *USDBRL) error {
	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	select {
	case <-ctxTimeout.Done():
		log.Println("[RegistrarCotacao] Requisição cancelada pelo servidor: Tempo de execução insuficiente")
	default:
	}

	db, err := gorm.Open(sqlite.Open("client_server_api.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	tx := db.WithContext(ctx)
	tx.AutoMigrate(&USDBRL{})
	tx.Create(cotacao)

	return nil
}
