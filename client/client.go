package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	errorHandler(err)

	res, err := http.DefaultClient.Do(req)
	errorHandler(err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	errorHandler(err)

	file, err := os.Create("cotacao.txt")
	errorHandler(err)

	fmt.Fprintf(file, "Dolar: {%s}", string(body))
}

func errorHandler(err error) {
	if err != nil {
		panic(err)
	}
}
