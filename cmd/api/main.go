package main

import (
	"github.com/example/api-service/internal/api"
	"github.com/example/api-service/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	e := gin.New()
	apiInstance := api.NewAPI(e, config.APIOptions.LoadConfig(""))
	apiInstance.Start()
}
