package storage

import (
	"context"

	"webcrawler/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	access     bool
	uri        string
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoDB(access bool, uri string) *MongoDB {
	return &MongoDB{
		access: access,
		uri:    uri,
	}
}

func (db *MongoDB) Connect() {
	if db.access {
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(db.uri))
		if err != nil {
			panic(err)
		}
		db.client = client
		db.collection = db.client.Database("webcrawler").Collection("pages")
		filter := bson.D{{}}
		// Deletes all documents in the collection
		db.collection.DeleteMany(context.TODO(), filter)
	}
}

func (db *MongoDB) Disconnect() {
	if db.access {
		db.client.Disconnect(context.TODO())
		db.access = false
	}
}

func (db *MongoDB) InsertPage(page models.Page) {
	if db.access {
		db.collection.InsertOne(context.TODO(), page)
	}
}
