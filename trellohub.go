package main

import (
	"github.com/vrutkovs/trellohub/pkg/gin"
)

func main() {
	r := gin.SetupGin()
	gin.StartGin(r)
}
