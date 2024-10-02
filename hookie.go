package main

import (
	"github.com/DeimosTech/hookie/db/mongo"
	_ "github.com/DeimosTech/hookie/internal/hook"
	"log/slog"
)

func init() {
	if mongo.GetDbConnection() == nil {
		panic("InitMongo should be called")
	}
	slog.Default().Info("hookiee in action")
}

func main() {}
