package main

import (
	"github.com/fredytarigan/Tendo/cmd"
	"github.com/fredytarigan/Tendo/pkg/tendo/config"
	"github.com/fredytarigan/Tendo/pkg/tendo/logger"
)

func init() {
	logger.Logger.Info("Starting application service")
	logger.Logger.Info("Initializing application config")
	
	config.SetConfigFile("./config")
}

func main() {
	command := cmd.NewCommandEngine();
	command.Run();
}