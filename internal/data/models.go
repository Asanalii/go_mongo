package data

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a movie that doesn't exist in our database.

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ItemConfilct      = errors.New("Item conflict")
)

// Create a Models struct which wraps the MovieModel
// kind of enveloping
type Models struct {
	Items  ItemModel
	Users  UserModel
	Tokens TokenModel
	Trades TradeModel
}

// method which returns a Models struct containing the initialized MovieModel.
func NewModels(db *mongo.Client) Models {
	return Models{
		Users:  UserModel{DB: db},
		Items:  ItemModel{DB: db},
		Trades: TradeModel{DB: db},
		// Tokens: TokenModel{DB: db},
	}

}
