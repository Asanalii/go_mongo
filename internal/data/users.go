package data

import (
	"context"
	"errors"
	"time"

	"github.com/shynggys9219/greenlight/internal/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	FirstName string `json:"firstname" bson:"firstname"`
	LastName  string `json:"lastname" bson:"lastname"`
	Email     string `json:"email" bson:"email"`
	Password  string `json:"password" bson:"password"`
}

type UserModel struct {
	DB *mongo.Client
}

var AnonymousUser = &User{}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.FirstName != "", "firstname", "must be provided")
	v.Check(len(user.FirstName) <= 500, "firstname", "must not be more than 500 bytes long")
	v.Check(user.LastName != "", "lastname", "must be provided")
	v.Check(len(user.LastName) <= 500, "lastname", "must not be more than 500 bytes long")
	// Call the standalone ValidateEmail() helper.
	ValidateEmail(v, user.Email)
	if user.Password != "" {
		ValidatePasswordPlaintext(v, user.Password)
	}
	// If the plaintext password is not nil, call the standalone
	// ValidatePasswordPlaintext() helper.

}

func (m UserModel) Insert(user *User) error {

	collection := m.DB.Database("nosql").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	result, _ := collection.InsertOne(ctx, user)

	defer cancel()

	if result == nil {
		return ErrEditConflict
	}

	return nil
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

func (m UserModel) GetByEmail(email string) (*User, error) {

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	collection := m.DB.Database("nosql").Collection("users")

	defer cancel()

	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	if err != nil {
		return nil, errors.New("Error with searching")
	}

	return &user, nil
}

func (p User) Matches(plaintextPassword string) (bool, error) {

	err := bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}
