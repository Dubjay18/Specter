package main

import (
	"net/http"
	"github.com/Dubjay/specter/internal/proxy"
)


func main() {
	proxy := proxy.New("http://localhost:5173")
	http.ListenAndServe(":8000", proxy)
}