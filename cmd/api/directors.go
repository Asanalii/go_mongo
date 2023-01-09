package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shynggys9219/greenlight/internal/data"
)

func (app *application) addDirectorHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name    string   `json:"name"`
		Surname string   `json:"surname"`
		Awards  []string `json:"awards"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	directors := &data.Directors{
		Name:    input.Name,
		Surname: input.Surname,
		Awards:  input.Awards,
	}

	fmt.Println(directors)

	err = app.models.Directors.Insert(directors)

	fmt.Println(directors)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/directors/%d", directors.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"director": directors}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showDirector(w http.ResponseWriter, r *http.Request) {

	name, err := app.readNameParam(r)

	if err != nil {
		app.notFoundResponse(w, r)
	}

	director, err := app.models.Directors.GetByName(name)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"director": director}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showDirectorsHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name       string
		Surname    string
		Awards     []string
		SearchName string
		data.Filters
	}

	// v := validator.New()

	qs := r.URL.Query()

	input.SearchName = app.readString(qs, "name", "")

	input.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "year", "awards", "-id", "-name", "-awards", "surname", "-surname"}

	directors, err := app.models.Directors.Get(input.SearchName, input.Filters)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send a JSON response containing the movie data.
	err = app.writeJSON(w, http.StatusOK, envelope{"directors": directors}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
