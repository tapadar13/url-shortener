//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	analyticsrepository "github.com/tapadar13/url-shortener/apps/api/internal/analytics/repository/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestAnalyticsRepositoryConcurrentDailyClicks(t *testing.T) {
	client := newIntegrationMongoClient(t)
	collection := client.AnalyticsCollection()
	repository := analyticsrepository.New(collection)
	ctx, cancel := context.WithTimeout(context.Background(), integrationTimeout)
	defer cancel()

	const clickCount = 32
	shortCode := "Ana1234"
	firstClickedAt := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	errors := make(chan error, clickCount)

	for offset := range clickCount {
		click, err := analytics.NewClick(shortCode, firstClickedAt.Add(time.Duration(offset)*time.Minute))
		if err != nil {
			t.Fatalf("create click %d: %v", offset, err)
		}

		go func() {
			errors <- repository.RecordClick(ctx, click)
		}()
	}

	for range clickCount {
		if err := <-errors; err != nil {
			t.Fatalf("record concurrent click: %v", err)
		}
	}

	filter := bson.D{{Key: "short_code", Value: shortCode}}
	documentCount, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		t.Fatalf("count daily analytics documents: %v", err)
	}
	if documentCount != 1 {
		t.Fatalf("expected one daily analytics document, got %d", documentCount)
	}

	var document struct {
		ShortCode      string    `bson:"short_code"`
		DayStart       time.Time `bson:"day_start"`
		ClickCount     int64     `bson:"click_count"`
		FirstClickedAt time.Time `bson:"first_clicked_at"`
		LastClickedAt  time.Time `bson:"last_clicked_at"`
	}
	if err := collection.FindOne(ctx, filter).Decode(&document); err != nil {
		t.Fatalf("read daily analytics document: %v", err)
	}

	expectedDayStart := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	expectedLastClickedAt := firstClickedAt.Add((clickCount - 1) * time.Minute)
	if document.ShortCode != shortCode ||
		!document.DayStart.Equal(expectedDayStart) ||
		document.ClickCount != clickCount ||
		!document.FirstClickedAt.Equal(firstClickedAt) ||
		!document.LastClickedAt.Equal(expectedLastClickedAt) {
		t.Fatalf("expected complete daily aggregate, got %+v", document)
	}
}
