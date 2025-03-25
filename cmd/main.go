package main

import (
	"flag"
	"github.com/GlebMoskalev/go-todo-api/internal/app"
	"log/slog"
	"os"
)

var flagConfig = flag.String("config", "./config/local.yaml", "path to the config file")

func main() {
	flag.Parse()

	if err := app.Run(*flagConfig); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}
