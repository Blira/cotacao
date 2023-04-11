package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
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

type ApiResponse struct {
	USDBRL Cotacao `json:"USDBRL"`
}

type CotacaoWithId struct {
	Id string `json:"id"`
	Cotacao
}

type ServerResponse struct {
	Cotacao float64 `json:"cotacao"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", FetchApi)
	http.ListenAndServe(":8080", mux)
}
func FetchApi(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("sqlite3", "./cotacao.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusGatewayTimeout)
		w.Write([]byte(`Internal Server Error`))
		return
	}
	defer res.Body.Close()

	var apiResponse ApiResponse
	err = json.NewDecoder(res.Body).Decode(&apiResponse)
	if err != nil {
		panic(err)
	}

	cotacao := apiResponse.USDBRL

	_, err = InsertCotacao(db, cotacao)
	if err != nil {
		panic(err)
	}

	value, err := strconv.ParseFloat(cotacao.Bid, 64)
	if err != nil {
		panic(err)
	}

	serverResponse := ServerResponse{Cotacao: value}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(serverResponse)
}

func InsertCotacao(db *sql.DB, cotacao Cotacao) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db.ExecContext(ctx, "create table cotacoes (id varchar(255),ask varchar(255),bid varchar(255),code varchar(255),codein varchar(255),createdate varchar(255),high varchar(255),low varchar(255),name varchar(255),pctchange varchar(255),timestamp varchar(255),varbid varchar(255), primary key (id))")

	stmt, err := db.PrepareContext(ctx, "insert into cotacoes (id,ask,bid,code,codein,createdate,high,low,name,pctchange,timestamp,varbid) values (?,?,?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	inserted, err := stmt.Exec(
		uuid.New().String(),
		cotacao.Ask,
		cotacao.Bid,
		cotacao.Code,
		cotacao.Codein,
		cotacao.CreateDate,
		cotacao.High,
		cotacao.Low,
		cotacao.Name,
		cotacao.PctChange,
		cotacao.Timestamp,
		cotacao.VarBid)
	if err != nil {
		return nil, err
	}
	return inserted, nil
}
