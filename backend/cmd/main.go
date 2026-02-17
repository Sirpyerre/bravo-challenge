package main

import "github.com/Sirpyerre/bravo-challenge/internal/config"

func main() {
	cfg := config.Load()
	srv := newServer(cfg)
	srv.start()
}
