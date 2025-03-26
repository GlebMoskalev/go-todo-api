package main

// @title Todo API
// @version 2.0
// @description This is a simple Todo API with authentication.
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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
