package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ServerResponse struct {
	Cotacao float64 `json:"cotacao"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("REQUEST FAILED WITH STATUS CODE: %d\n", res.StatusCode)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var serverResponse ServerResponse
	err = json.Unmarshal(body, &serverResponse)
	if err != nil {
		println(err)
	}

	content := "Dólar:" + strconv.FormatFloat(serverResponse.Cotacao, 'f', 4, 64)
	filename := "cotacao.txt"

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		panic(err)
	}

}
