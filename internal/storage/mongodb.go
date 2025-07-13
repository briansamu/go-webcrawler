package storage

import (
	"context"
	"fmt"
	"sort"
	"strings"

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
		fmt.Println("Database cleared - all previous pages deleted")
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
		_, err := db.collection.InsertOne(context.TODO(), page)
		if err != nil {
			fmt.Printf("Error inserting page %s: %v\n", page.Url, err)
		} else {
			fmt.Printf("Successfully inserted page: %s\n", page.Url)
		}
	} else {
		fmt.Printf("Database not accessible, cannot insert page: %s\n", page.Url)
	}
}

func (db *MongoDB) SearchPages(query string, page, limit int) ([]models.Page, int, error) {
	if !db.access {
		return nil, 0, fmt.Errorf("database not accessible")
	}

	ctx := context.TODO()

	// Create text search filter
	filter := bson.M{
		"$or": []bson.M{
			{"title": bson.M{"$regex": query, "$options": "i"}},
			{"content": bson.M{"$regex": query, "$options": "i"}},
			{"url": bson.M{"$regex": query, "$options": "i"}},
		},
	}

	fmt.Printf("Searching for query: '%s'\n", query)
	fmt.Printf("Search filter: %+v\n", filter)

	// Count total results
	total, err := db.collection.CountDocuments(ctx, filter)
	if err != nil {
		fmt.Printf("Error counting documents: %v\n", err)
		return nil, 0, err
	}

	fmt.Printf("Total search results found: %d\n", total)

	// If no results, let's check what data exists in the database
	if total == 0 {
		fmt.Println("No search results found. Checking database content...")

		// Get a sample document to see the actual field structure
		var sampleDoc bson.M
		err = db.collection.FindOne(ctx, bson.M{}).Decode(&sampleDoc)
		if err != nil {
			fmt.Printf("Error getting sample document: %v\n", err)
		} else {
			fmt.Printf("Sample document structure: %+v\n", sampleDoc)
		}
	}

	// Get all matching results (without pagination for scoring)
	cursor, err := db.collection.Find(ctx, filter)
	if err != nil {
		fmt.Printf("Error executing search: %v\n", err)
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var allPages []models.Page
	err = cursor.All(ctx, &allPages)
	if err != nil {
		fmt.Printf("Error decoding search results: %v\n", err)
		return nil, 0, err
	}

	// Calculate scores for each page
	for i := range allPages {
		allPages[i].Score = db.calculateScore(allPages[i], query)
	}

	// Sort by score (highest first)
	sort.Slice(allPages, func(i, j int) bool {
		return allPages[i].Score > allPages[j].Score
	})

	// Apply pagination after sorting
	start := (page - 1) * limit
	end := start + limit

	if start >= len(allPages) {
		return []models.Page{}, int(total), nil
	}

	if end > len(allPages) {
		end = len(allPages)
	}

	paginatedPages := allPages[start:end]

	fmt.Printf("Search returned %d pages\n", len(paginatedPages))
	return paginatedPages, int(total), nil
}

func (db *MongoDB) calculateScore(page models.Page, query string) float64 {
	queryLower := strings.ToLower(query)
	titleLower := strings.ToLower(page.Title)
	contentLower := strings.ToLower(page.Content)
	urlLower := strings.ToLower(page.Url)

	score := 0.0

	// Title matches (highest weight)
	if strings.Contains(titleLower, queryLower) {
		score += 10.0

		// Bonus for exact title match
		if titleLower == queryLower {
			score += 20.0
		}

		// Bonus for title starting with query
		if strings.HasPrefix(titleLower, queryLower) {
			score += 5.0
		}

		// Count occurrences in title
		titleOccurrences := strings.Count(titleLower, queryLower)
		score += float64(titleOccurrences) * 3.0
	}

	// URL matches (medium weight)
	if strings.Contains(urlLower, queryLower) {
		score += 5.0

		// Bonus for domain name match
		if strings.Contains(urlLower, queryLower) {
			score += 2.0
		}

		// Count occurrences in URL
		urlOccurrences := strings.Count(urlLower, queryLower)
		score += float64(urlOccurrences) * 2.0
	}

	// Content matches (lower weight)
	if strings.Contains(contentLower, queryLower) {
		score += 1.0

		// Count occurrences in content
		contentOccurrences := strings.Count(contentLower, queryLower)
		score += float64(contentOccurrences) * 0.5

		// Bonus for content starting with query
		if strings.HasPrefix(contentLower, queryLower) {
			score += 2.0
		}
	}

	// Length penalty (shorter content is more relevant)
	if len(page.Content) > 0 {
		lengthPenalty := float64(len(page.Content)) / 10000.0
		score -= lengthPenalty
	}

	// Ensure minimum score of 0
	if score < 0 {
		score = 0
	}

	return score
}

func (db *MongoDB) GetPages(page, limit int) ([]models.Page, int, error) {
	if !db.access {
		return nil, 0, fmt.Errorf("database not accessible")
	}

	ctx := context.TODO()

	// Count total documents
	total, err := db.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, 0, err
	}

	// Set up pagination
	skip := (page - 1) * limit
	opts := options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)).SetSort(bson.M{"_id": -1})

	// Get pages
	cursor, err := db.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var pages []models.Page
	err = cursor.All(ctx, &pages)
	if err != nil {
		return nil, 0, err
	}

	return pages, int(total), nil
}

func (db *MongoDB) GetTotalPages() (int, error) {
	if !db.access {
		return 0, fmt.Errorf("database not accessible")
	}

	count, err := db.collection.CountDocuments(context.TODO(), bson.M{})
	return int(count), err
}
