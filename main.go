package main

import (
	"fmt"

	"github.com/fredytarigan/Tendo/cmd"
	"github.com/fredytarigan/Tendo/pkg/tendo/config"
)

func init() {
	fmt.Println("Starting application service");
	fmt.Println("Initializing application config");
	config.SetConfigFile("./config")
}

func main() {
	command := cmd.NewCommandEngine();
	command.Run();
}