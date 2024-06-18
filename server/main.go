package main

import (
	"fmt"
	"net/http"

	"github.com/kidommoc/gustrody/internal/config"
	"github.com/kidommoc/gustrody/internal/db"
	"github.com/kidommoc/gustrody/internal/models"
	"github.com/kidommoc/gustrody/internal/router"
)

func main() {
	config.Get()
	db.Init()
	models.Init()

	router := router.Router()
	srv := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil {
		fmt.Errorf("Server shutdown... %s", err.Error())
	}
}
