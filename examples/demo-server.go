package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		// Simulate auth endpoint
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "token": "abc"})
	})

	http.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		// Sometimes fails when auth is slow; we'll just return ok
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "signed"})
	})

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		// return success
		fmt.Fprintln(w, "orders list")
	})

	fmt.Println("demo backend listening :3000")
	http.ListenAndServe(":3000", nil)
}
