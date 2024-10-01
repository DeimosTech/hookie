package main

import (
	_ "github.com/DeimosTech/hookie/internal/hook"
	"log/slog"
)

func init() {
	slog.Default().Info("hookiee in action")
}

func main() {

}
