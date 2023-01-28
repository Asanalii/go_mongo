package data

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// By default, the keys in the JSON object are equal to the field names in the struct ( ID,
// CreatedAt, Title and so on).
type Item struct {
	ID          int64     `json:"id" bson:"id"`
	User_Email  string    `json:"user_email" bson:"user_email"`   // Unique integer ID for the movie
	CreatedAt   time.Time `json:"-" bson:"-"`                     // Timestamp for when the movie is added to our database, "-" directive, hidden in response
	Name        string    `json:"name" bson:"name"`               // Movie title
	Description string    `json:"description" bson:"description"` // Movie title
	Status      string    `json:"status" bson:"status"`
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type ItemModel struct {
	DB *mongo.Client
}

// method for inserting a new record in the movies table.
func (m ItemModel) Insert(item *Item) error {

	// collection := m.DB.Database("nosql").Collection("items")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	counterCollection := m.DB.Database("nosql").Collection("counters")
	collection_update := m.DB.Database("nosql").Collection("users")

	filter := bson.M{"_id": 0}
	update := bson.M{"$inc": bson.M{"seq": 1}}

	opts := options.FindOneAndUpdate().SetUpsert(true)

	var counter struct {
		ID  int64 `bson:"_id"`
		Seq int64 `bson:"seq"`
	}

	err := counterCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&counter)

	// ВОТ ТУТ ОШИБКА ВЫХОДИТ КОГДА СОЗДАЮ ПРОДУКТ ЗАНОВО
	if err != nil {
		// log.Fatal(err)
		fmt.Println(err)
	}

	// fmt.Println("New ID:", counter.Seq)

	item.ID = counter.Seq
	// result, _ := collection.InsertOne(ctx, item)

	filter2 := bson.D{{"email", item.User_Email}}
	update2 := bson.D{{"$push", bson.D{{"items", item}}}}

	_, err = collection_update.UpdateOne(ctx, filter2, update2)

	defer cancel()

	return nil
	// return m.DB.QueryRow(query, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres)).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m ItemModel) GetById(id int64) (*Item, error) {

	var item Item

	collection := m.DB.Database("nosql").Collection("users")

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// fmt.Println(strconv.FormatInt(id, 10) + "1111")
	pipeline := []bson.M{
		{"$unwind": "$items"},
		{"$match": bson.M{"items.id": id}},
		{"$replaceRoot": bson.M{"newRoot": "$items"}},
		{"$project": bson.M{"_id": 0}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	var result bson.M

	for cursor.Next(context.TODO()) {
		err := cursor.Decode(&result)
		if err != nil {
			panic(err)
		}
		config := &mapstructure.DecoderConfig{
			Result:  &item,
			TagName: "bson",
		}
		decoder, err := mapstructure.NewDecoder(config)
		if err != nil {
			panic(err)
		}
		decoder.Decode(result)
	}

	if err != nil {
		return nil, errors.New("Error with searching")
	}

	return &item, nil
}

func (m ItemModel) Get() ([]*Item, error) {

	collection := m.DB.Database("nosql").Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pipeline := []bson.M{
		{"$unwind": "$items"},
		{"$match": bson.M{"items.status": "available"}},

		{"$group": bson.M{
			"_id":      nil,
			"allItems": bson.M{"$push": "$items"},
		}},
		{"$project": bson.M{"_id": 0}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Fatal(err)
	}

	// items := []*Item{}

	var result bson.M
	var allItems []*Item

	for cursor.Next(context.TODO()) {
		err := cursor.Decode(&result)
		if err != nil {
			fmt.Println(err)
		}
		for _, item := range result["allItems"].(primitive.A) {
			var itemData Item
			bsonBytes, _ := bson.Marshal(item.(primitive.M))
			bson.Unmarshal(bsonBytes, &itemData)
			allItems = append(allItems, &itemData)
		}
	}

	return allItems, nil
}

func (m ItemModel) Update(item *Item) error {

	collection := m.DB.Database("nosql").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{
		"email": item.User_Email,
		"items": bson.M{
			"$elemMatch": bson.M{
				"id": item.ID,
			},
		},
	}
	update := bson.M{
		"$set": bson.M{
			"items.$.name":        item.Name,
			"items.$.description": item.Description,
			"items.$.status":      item.Status,
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (m ItemModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	collection := m.DB.Database("nosql").Collection("users")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	delete := bson.M{
		"$pull": bson.M{
			"items": bson.M{
				"id": id,
			},
		},
	}

	filter := bson.M{
		"items.id": id,
	}

	_, err := collection.UpdateOne(ctx, filter, delete)

	if err != nil {
		return err
	}

	return nil
}
