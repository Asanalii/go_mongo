package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/shynggys9219/greenlight/internal/data"
	"github.com/shynggys9219/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

// func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
// 	// Create an anonymous struct to hold the expected data from the request body.
// var input struct {
// 	Name     string `json:"name"`
// 	Email    string `json:"email"`
// 	Password string `json:"password"`
// }
// 	// Parse the request body into the anonymous struct.
// 	err := app.readJSON(w, r, &input)
// 	if err != nil {
// 		app.badRequestResponse(w, r, err)
// 		return
// 	}
// user := &data.User{
// 	Name:      input.Name,
// 	Email:     input.Email,
// 	Activated: false,
// }
// 	// Use the Password.Set() method to generate and store the hashed and plaintext
// 	// passwords.
// err = user.Password.Set(input.Password)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	v := validator.New()
// 	// Validate the user struct and return the error messages to the client if any of
// 	// the checks fail.
// 	if data.ValidateUser(v, user); !v.Valid() {
// 		app.failedValidationResponse(w, r, v.Errors)
// 		return
// 	}
// 	// Insert the user data into the database.
// 	err = app.models.Users.Insert(user)
// 	if err != nil {
// 		switch {
// 		// If we get a ErrDuplicateEmail error, use the v.AddError() method to manually
// 		// add a message to the validator instance, and then call our
// 		// failedValidationResponse() helper.
// 		case errors.Is(err, data.ErrDuplicateEmail):
// 			v.AddError("email", "a user with this email address already exists")
// 			app.failedValidationResponse(w, r, v.Errors)
// 		default:
// 			app.serverErrorResponse(w, r, err)
// 		}
// 		return
// 	}

// 	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	data := map[string]any{
// 		"activationToken": token.Plaintext,
// 		"userID":          user.ID,
// 	}
// 	// Send the welcome email, passing in the map above as dynamic data.
// 	err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 		return
// 	}

// 	// Write a JSON response containing the user data along with a 201 Created status
// 	// code.
// 	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
// 	if err != nil {
// 		app.serverErrorResponse(w, r, err)
// 	}
// }

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName string `json:"firstname bson:"firstname"`
		LastName  string `json:"lastname bson:"lastname"`
		Email     string `json:"email" bson:"email"`
		Password  string `json:"password" bson:"password"`
	}
	user := &data.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
	}

	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// 	return
	// }
	json.NewDecoder(r.Body).Decode(&user)

	v := validator.New()

	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user.Password = getHash([]byte(user.Password))

	err := app.models.Users.Insert(user)

	if err != nil {
		switch {
		// If we get a ErrDuplicateEmail error, use the v.AddError() method to manually
		// add a message to the validator instance, and then call our
		// failedValidationResponse() helper.
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// json.NewEncoder(w).Encode(err)

	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func getHash(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
