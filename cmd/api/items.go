package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shynggys9219/greenlight/internal/data"
)

func (app *application) createItemHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Name        string `json:"name" bson:"name"`
		Status      string `json:"status" bson:"status"`
		Description string `json:"description" bson:"description"`
	}

	err := app.readJSON(w, r, &input)
	user := app.contextGetUser(r)

	//Statuses:
	//Available - common status
	//Trading - when item is trading
	//Deleted - when item is deleted or made trade

	item := &data.Item{
		Name:        input.Name,
		Status:      "available",
		Description: input.Description,
		User_Email:  user.Email,
	}

	err = app.models.Items.Insert(item)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/items/%d", item.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"item": item}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	// // Dump the contents of the input struct in a HTTP response.
	// fmt.Fprintf(w, "%+v\n", input) //+v here is adding the field name of a value // https://pkg.go.dev/fmt
}

func (app *application) showItemHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	// Create a new instance of the Movie struct, containing the ID we extracted from
	// the URL and some dummy data. Also notice that we deliberately haven't set a
	// value for the Year field.

	item, err := app.models.Items.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}
	// Encode the struct to JSON and send it as the HTTP response.
	// using envelope
	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) showItemsHandler(w http.ResponseWriter, r *http.Request) {

	// movie, err := app.models.Movies.Get()

	// var input struct {
	// 	Title  string
	// 	Genres []string
	// 	data.Filters
	// }

	// v := validator.New()

	// qs := r.URL.Query()

	// input.Title = app.readString(qs, "title", "")
	// input.Genres = app.readCSV(qs, "genres", []string{})

	// input.Filters.Page = app.readInt(qs, "page", 1)
	// input.Filters.PageSize = app.readInt(qs, "page_size", 20)

	// input.Sort = app.readString(qs, "sort", "id")
	// input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	items, err := app.models.Items.Get()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send a JSON response containing the movie data.
	err = app.writeJSON(w, http.StatusOK, envelope{"items": items}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

//UPDATE STATUS надо добавить

func (app *application) updateItemHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	item, err := app.models.Items.GetById(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	var input struct {
		Name        *string `json:"name" bson:"name"`
		Description *string `json:"description" bson:"description"`
		Status      *string `json:"status" bson:"status"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if input.Name != nil {
		item.Name = *input.Name
	}
	if input.Description != nil {
		item.Description = *input.Description
	}

	if input.Status != nil {
		item.Status = *input.Status
	}

	err = app.models.Items.Update(item)

	err = app.writeJSON(w, http.StatusOK, envelope{"item": item}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteItemHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	user := app.contextGetUser(r)

	res, err := app.models.Items.GetById(id)

	if user.Email != res.User_Email {
		app.invalidCredentialsResponse(w, r)
		// app.anotherUserResponse(w, r)
		return
	}

	item := app.models.Items.Delete(id)
	if item != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	item = app.writeJSON(w, http.StatusOK, envelope{"item": "Deleted successfully"}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
