package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// Initialize a new httprouter router instance.

	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.requireActivatedUser(app.healthcheckHandler))
	// router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	// router.HandlerFunc(http.MethodGet, "/v1/movies", app.showMoviesHandler)
	// router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	// router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	// router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)

	router.HandlerFunc(http.MethodPost, "/v1/items", app.requireActivatedUser(app.createItemHandler))
	router.HandlerFunc(http.MethodGet, "/v1/item/:id", app.requireActivatedUser(app.showItemHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/item/:id", app.requireActivatedUser(app.updateItemHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/item/:id", app.requireActivatedUser(app.deleteItemHandler))
	router.HandlerFunc(http.MethodGet, "/v1/items", app.showItemsHandler)

	router.HandlerFunc(http.MethodPost, "/v1/trades", app.requireActivatedUser(app.createTradeHandler))
	router.HandlerFunc(http.MethodGet, "/v1/trade/:id", app.requireActivatedUser(app.showTradeHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/trade/:id/accept", app.requireActivatedUser(app.acceptTradeHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/trade/:id/decline", app.requireActivatedUser(app.declineTradeHandler))

	// удаление трейда поидеи ненужная штука
	// router.HandlerFunc(http.MethodDelete, "/v1/trade/:id", app.requireActivatedUser(app.deleteTradeHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler) //login
	// router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
	// return app.recoverPanic(router)

}
