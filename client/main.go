package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx := context.Background()

	cambio, err := SolicitarCotacaoDolar(ctx)
	if err != nil {
		panic(err)
	}

	err = SalvarCotacaoTxt(*cambio)
	if err != nil {
		panic(err)
	}
}

func SolicitarCotacaoDolar(ctx context.Context) (*string, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	select {
	case <-ctxTimeout.Done():
		log.Println("[SolicitarCotacaoDolar] Requisição cancelada pelo cliente: Tempo de execução insuficiente")
	default:
	}

	req, err := http.Get("http://localhost:8080/cotacao")
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var cambio string
	err = json.Unmarshal(body, &cambio)
	if err != nil {
		return nil, err
	}

	return &cambio, nil
}

func SalvarCotacaoTxt(cambio string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}

	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dólar: %s", cambio))
	if err != nil {
		return err
	}

	return nil
}
