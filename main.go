package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseConnection struct {
	access     bool
	uri        string
	client     *mongo.Client
	collection *mongo.Collection
}

func (db *DatabaseConnection) connect() {
	if db.access {
		db.uri = os.Getenv("MONGO_URI")
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

func (db *DatabaseConnection) disconnect() {
	if db.access {
		db.client.Disconnect(context.TODO())
		db.access = false
	}
}

func (db *DatabaseConnection) insertPage(page Page) {
	if db.access {
		db.collection.InsertOne(context.TODO(), page)
	}
}

type Page struct {
	Url     string
	Title   string
	Content string
}

func fetchPage(url string) (string, string, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var title, content string

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(1500*time.Millisecond), // Wait for hydration
		chromedp.Title(&title),
		chromedp.OuterHTML(`html`, &content, chromedp.ByQuery),
	)

	return title, content, err
}

func main() {
	dbAccess := true

	if godotenv.Load() != nil {
		fmt.Println("Error loading .env file")
		dbAccess = false
	}

	db := DatabaseConnection{access: dbAccess, uri: "", client: nil, collection: nil}
	db.connect()

	seed := os.Getenv("SEED_URL")
	title, content, err := fetchPage(seed)
	if err != nil {
		fmt.Println("Error fetching page:", err)
		return
	}

	db.insertPage(Page{Url: seed, Title: title, Content: content})

	db.disconnect()
}

// Add page to queue
// Get page body
// Parse page body for links
// Add links to queue
// Add page to visited
// Repeat until no more pages to visit
