package main

import (
	"fmt"
	"github.com/BetelgeuseTb/betelgeuse-orbitum/pkg/utils/reader"
)

const (
	LOGO_PATH = "internal/resources/logo.txt"
)

func main() {
	data, _ := reader.NewFileReader().ReadFile(LOGO_PATH)
	fmt.Println(string(data))
}
