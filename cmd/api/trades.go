package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/shynggys9219/greenlight/internal/data"
)

func (app *application) createTradeHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Giver_ID    int64 `json:"giver_id" bson:"giver_id"`
		Receiver_ID int64 `json:"receiver_id" bson:"receiver_id"`
	}

	err := app.readJSON(w, r, &input)

	user := app.contextGetUser(r)

	giver, err := app.models.Items.GetById(input.Giver_ID)
	receiver, err := app.models.Items.GetById(input.Receiver_ID)

	if user.Email != giver.User_Email {
		app.invalidCredentialsResponse(w, r)
		return
	}

	if user.Email == receiver.User_Email {
		app.errorResponse(w, r, http.StatusMethodNotAllowed, "You can not make trade to your own item")
		return
	}

	giver.Status = "trading"
	receiver.Status = "trading"

	trade := &data.Trade{
		Giver:    *giver,
		Reciever: *receiver,
	}

	err = app.models.Items.Update(giver)
	err = app.models.Items.Update(receiver)

	//waiting_action = ждёт экшн юзера
	//canceled_trade = отменённый
	//confirmed_trade = принятый

	trade.Status = "waiting_action"
	err = app.models.Trades.MakeTrade(trade)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/trades/%d", trade.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"trade": trade}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) acceptTradeHandler(w http.ResponseWriter, r *http.Request) {

	trade_id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	user := app.contextGetUser(r)

	trade, err := app.models.Trades.GetById(trade_id)
	giver, err := app.models.Items.GetById(trade.Giver.ID)
	receiver, err := app.models.Items.GetById(trade.Reciever.ID)

	if user.Email == trade.Giver.User_Email {
		err = app.models.Trades.UpdateGiver(trade_id, trade.Giver.ID, "Accepted")
		trade.Giver.Status = "Accepted"
	} else if user.Email == trade.Reciever.User_Email {
		err = app.models.Trades.UpdateReceiver(trade_id, trade.Giver.ID, "Accepted")
		trade.Reciever.Status = "Accepted"
	} else {
		app.invalidCredentialsResponse(w, r)
		return
	}

	if trade.Giver.Status == "Accepted" && trade.Reciever.Status == "Accepted" {
		err = app.models.Trades.UpdateTraded(trade.ID)

		giver.Status = "deleted"
		receiver.Status = "deleted"
		err = app.models.Items.Update(giver)
		err = app.models.Items.Update(receiver)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"trade": "You succefully accepted!"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) declineTradeHandler(w http.ResponseWriter, r *http.Request) {

	trade_id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	user := app.contextGetUser(r)

	trade, err := app.models.Trades.GetById(trade_id)
	giver, err := app.models.Items.GetById(trade.Giver.ID)
	receiver, err := app.models.Items.GetById(trade.Reciever.ID)

	if user.Email == trade.Giver.User_Email {
		err = app.models.Trades.UpdateGiver(trade_id, trade.Giver.ID, "Declined")
		trade.Giver.Status = "Declined"
	} else if user.Email == trade.Reciever.User_Email {
		err = app.models.Trades.UpdateReceiver(trade_id, trade.Giver.ID, "Declined")
		trade.Reciever.Status = "Declined"
	} else {
		app.invalidCredentialsResponse(w, r)
		return
	}

	if trade.Giver.Status == "Declined" || trade.Reciever.Status == "Declined" {
		err = app.models.Trades.UpdateDeclined(trade.ID)
		giver.Status = "available"
		receiver.Status = "available"
		err = app.models.Items.Update(giver)
		err = app.models.Items.Update(receiver)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"trade": "You succefully declined!"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// acceptTradeHandler

func (app *application) showTradeHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	// Create a new instance of the Movie struct, containing the ID we extracted from
	// the URL and some dummy data. Also notice that we deliberately haven't set a
	// value for the Year field.

	trade, err := app.models.Trades.GetById(id)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"trade": trade}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteTradeHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	trade, err := app.models.Trades.GetById(id)

	giver, err := app.models.Items.GetById(trade.Giver.ID)
	receiver, err := app.models.Items.GetById(trade.Reciever.ID)

	giver.Status = "available"
	receiver.Status = "available"
	err = app.models.Items.Update(giver)
	err = app.models.Items.Update(receiver)

	item := app.models.Trades.Delete(id)
	if item != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	item = app.writeJSON(w, http.StatusOK, envelope{"trade": "Deleted successfully"}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
