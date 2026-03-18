package main

import (
    "fmt"
    "net/http"
    "os"
)

func main() {
    port := "3000"
    if len(os.Args) > 1 {
        port = os.Args[1]
    }
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, `{"server":"live","port":"%s","path":"%s"}`, port, r.URL.Path)
    })
    fmt.Printf("test server on :%s\n", port)
    http.ListenAndServe(":"+port, nil)
}