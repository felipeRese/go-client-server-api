package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ServerResponse struct {
	Bid float64 `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Printf("Erro ao criar requisição: %v\n", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("Erro: o tempo de execução foi insuficiente (timeout).")
		} else {
			fmt.Printf("Erro ao realizar requisição: %v\n", err)
		}
		return
	}
	defer res.Body.Close()

	var data ServerResponse
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		fmt.Printf("Erro ao decodificar resposta: %v\n", err)
		return
	}

	file, err := os.Create("cotacao.txt")
	if err != nil {
		fmt.Printf("Erro ao criar arquivo: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %f\n", data.Bid))
	if err != nil {
		fmt.Printf("Erro ao escrever no arquivo: %v\n", err)
		return
	}

	fmt.Println("Cotação salva com sucesso em 'cotacao.txt'")
}
