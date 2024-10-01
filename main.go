package main

import (
	_ "github.com/DeimosTech/hookie/internal/hook"
	"log/slog"
)

func main() {
	slog.Default().Info("hookiee in action")
}
