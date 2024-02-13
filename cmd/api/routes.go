package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.NotFound(app.notFound)
	mux.MethodNotAllowed(app.methodNotAllowed)

	mux.Use(app.logAccess)
	mux.Use(app.recoverPanic)

	v1Router := chi.NewRouter()

	v1Router.Get("/v1/healthcheck", app.healthcheckHandler)

	// mux.Group(func(mux chi.Router) {
	// 	mux.Use(app.authenticate)
	// 	mux.Use(app.requireAuthenticatedUser)
	//
	// 	mux.Get("/protected", app.protected)
	// })

	mux.Mount("/api", v1Router)

	return mux
}
