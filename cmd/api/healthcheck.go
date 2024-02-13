package main

import (
	"fmt"
	"net/http"

	"github.com/brGuirra/uai/internal/response"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Caiu aqui")

	data := map[string]any{
		"status": "available",
		"systemInfo": map[string]string{
			"enviroment": app.config.env,
			"version":    version,
		},
	}

	err := response.JSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}
