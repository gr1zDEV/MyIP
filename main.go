package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

func ipHandler(w http.ResponseWriter, r *http.Request) {
	ip := r.Header.Get("X-Forwarded-For")
	
	if strings.Contains(ip, ",") {
		ip = strings.Split(ip, ",")[0]
	}

	if ip == "" {
		ip = r.RemoteAddr
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, ip)
}

func main() {
	http.HandleFunc("/", ipHandler)
	
	fmt.Println("Starting server on port 8000...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}
