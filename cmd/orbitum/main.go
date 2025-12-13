package main

import (
	"fmt"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils/logger"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/internal/utils/reader"
)

const (
	LOGO_PATH = "internal/resources/logo.txt"
)

func main() {
	data, _ := reader.NewFileReader().ReadFile(LOGO_PATH)
	fmt.Println(string(data))
	logger.LogInfo("Starting application ...", "main")
	logger.LogInfo("Application finished.", "main")
}
