package data

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Trade struct {
	ID       int64  `json:"_id" bson:"_id"`
	Giver    Item   `json:"giver" bson:"giver"`
	Reciever Item   `json:"receiver" bson:"receiver"`
	Status   string `json:"status" bson:"status"`
}

// Define a MovieModel struct type which wraps a sql.DB connection pool.
type TradeModel struct {
	DB *mongo.Client
}

func (m TradeModel) MakeTrade(trade *Trade) error {

	// collection := m.DB.Database("nosql").Collection("trades")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	counterCollection := m.DB.Database("nosql").Collection("counters_trade")
	collection_update := m.DB.Database("nosql").Collection("trades")

	filter := bson.M{"_id": 0}
	update := bson.M{"$inc": bson.M{"seq": 1}}

	opts := options.FindOneAndUpdate().SetUpsert(true)

	var counter struct {
		ID  int64 `bson:"_id"`
		Seq int64 `bson:"seq"`
	}

	err := counterCollection.FindOneAndUpdate(context.TODO(), filter, update, opts).Decode(&counter)

	if err != nil {
		fmt.Println(err)
	}

	trade.ID = counter.Seq

	// filter2 := bson.D{{"trades", trade}}

	// fmt.Println(filter2)
	// update2 := bson.D{{"$push", bson.D{{"items", item}}}}

	result, _ := collection_update.InsertOne(ctx, trade)
	defer cancel()

	if result == nil {
		return ErrEditConflict
	}

	return nil
}

func (m TradeModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	collection := m.DB.Database("nosql").Collection("trades")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// update := bson.M{
	// 	"$pull": bson.M{
	// 		"items": bson.M{
	// 			"id": id,
	// 		},
	// 	},
	// }

	filter := bson.M{
		"_id": id,
	}

	// _, err := collection.UpdateOne(ctx, filter, update)

	_, err := collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	return nil
}

func (m TradeModel) GetById(id int64) (*Trade, error) {

	var trade Trade

	collection := m.DB.Database("nosql").Collection("trades")

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&trade)

	if err != nil {
		return nil, errors.New("Error with searching")
	}

	return &trade, nil
}

func (m TradeModel) UpdateTraded(id_trade int64) error {

	collection := m.DB.Database("nosql").Collection("trades")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": id_trade,
	}
	update := bson.M{
		"$set": bson.M{
			"status": "Confirmed",
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)

	return err
}

func (m TradeModel) UpdateDeclined(id_trade int64) error {

	collection := m.DB.Database("nosql").Collection("trades")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": id_trade,
	}
	update := bson.M{
		"$set": bson.M{
			"status": "Cancelled",
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)

	return err
}

func (m TradeModel) UpdateGiver(id_trade int64, id_user int64, status string) error {

	collection := m.DB.Database("nosql").Collection("trades")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": id_trade,
	}
	update := bson.M{
		"$set": bson.M{
			"giver.status": status,
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)

	return err
}

func (m TradeModel) UpdateReceiver(id_trade int64, id_user int64, status string) error {

	collection := m.DB.Database("nosql").Collection("trades")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": id_trade,
	}

	update := bson.M{
		"$set": bson.M{
			"receiver.status": status,
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)

	return err
}
