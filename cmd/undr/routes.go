package main

import (
	"fmt"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("GET /health/live", live)

	return router
}

func live(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "alive and rocking")
}
