package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type USDBRL struct {
	Data ExchangeRate `json:"USDBRL"`
}

type ExchangeRate struct {
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

type Exchange struct {
	ID       string
	Name     string
	Exchange string
}

func NewExchange(name string, exchange string) *Exchange {
	return &Exchange{
		ID:       uuid.NewString(),
		Name:     name,
		Exchange: exchange,
	}
}

func main() {
	http.HandleFunc("/cotacao", handler)

	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("Request Iniciada")
	defer log.Println("Request Finalizada")

	db, err := sql.Open("sqlite3", "./database.sqlite")
	errorHandler(err)
	defer db.Close()

	sts := `CREATE TABLE IF NOT EXISTS exchanges(id VARCHAR PRIMARY KEY, name TEXT, exchange VARCHAR);`
	_, err = db.Exec(sts)
	errorHandler(err)

	output, err := GetDolarExchangeRate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	exchange := NewExchange(output.Data.Name, output.Data.Bid)

	err = insertProduct(db, *exchange)
	errorHandler(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(output.Data.Bid)

}

func GetDolarExchangeRate() (*USDBRL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

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

	var e USDBRL
	err = json.Unmarshal(body, &e)
	if err != nil {
		return nil, err
	}

	return &e, nil

}

func insertProduct(db *sql.DB, data Exchange) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.PrepareContext(ctx, "insert into exchanges(id, name, exchange) values(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(data.ID, data.Name, data.Exchange)
	if err != nil {
		return err
	}
	return nil
}

func errorHandler(err error) {
	if err != nil {
		panic(err)
	}
}
