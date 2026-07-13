package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/ratelimit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRepositoryIncrementAtomicallyUpdatesCounter(t *testing.T) {
	t.Parallel()

	windowStart := time.Date(2026, 7, 13, 10, 5, 0, 123456789, time.FixedZone("IST", 5*60*60+30*60))
	expiresAt := windowStart.Add(time.Minute)
	collection := &fakeCollection{
		result: mongo.NewSingleResultFromDocument(counterDocument{Count: 2}, nil, nil),
	}
	repository := newRepository(collection)

	count, err := repository.Increment(context.Background(), ratelimit.IncrementParams{
		ClientKey:   " 203.0.113.10 ",
		WindowStart: windowStart,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		t.Fatalf("expected counter increment to succeed: %v", err)
	}

	if count != 2 {
		t.Fatalf("expected count 2, got %d", count)
	}

	assertCounterFilter(t, collection.filter, "203.0.113.10", windowStart.UTC())
	assertCounterUpdate(t, collection.update, expiresAt.UTC())

	updateOptions := findOneAndUpdateOptions(t, collection.options)
	if updateOptions.Upsert == nil || !*updateOptions.Upsert {
		t.Fatalf("expected upsert option, got %#v", updateOptions.Upsert)
	}

	if updateOptions.ReturnDocument == nil || *updateOptions.ReturnDocument != options.After {
		t.Fatalf("expected updated document to be returned, got %#v", updateOptions.ReturnDocument)
	}
}

func TestRepositoryIncrementRejectsNilCollection(t *testing.T) {
	t.Parallel()

	_, err := New(nil).Increment(context.Background(), validIncrementParams())
	if err == nil || !strings.Contains(err.Error(), "collection") {
		t.Fatalf("expected collection error, got %v", err)
	}
}

func TestRepositoryIncrementValidatesParametersBeforeQuerying(t *testing.T) {
	t.Parallel()

	collection := &fakeCollection{}
	repository := newRepository(collection)

	_, err := repository.Increment(context.Background(), ratelimit.IncrementParams{})
	if !errors.Is(err, ratelimit.ErrClientKeyRequired) {
		t.Fatalf("expected client key error, got %v", err)
	}

	if collection.called {
		t.Fatal("expected collection not to be called")
	}
}

func TestRepositoryIncrementRejectsInvalidExpiration(t *testing.T) {
	t.Parallel()

	params := validIncrementParams()
	params.ExpiresAt = params.WindowStart
	repository := newRepository(&fakeCollection{})

	_, err := repository.Increment(context.Background(), params)
	if !errors.Is(err, ratelimit.ErrExpirationInvalid) {
		t.Fatalf("expected expiration error, got %v", err)
	}
}

func TestRepositoryIncrementWrapsDatabaseError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	repository := newRepository(&fakeCollection{
		result: mongo.NewSingleResultFromDocument(bson.D{}, expectedErr, nil),
	})

	_, err := repository.Increment(context.Background(), validIncrementParams())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected database error, got %v", err)
	}
}

func assertCounterFilter(t *testing.T, value any, clientKey string, windowStart time.Time) {
	t.Helper()

	filter, ok := value.(bson.D)
	if !ok || len(filter) != 1 || filter[0].Key != "_id" {
		t.Fatalf("expected counter ID filter, got %#v", value)
	}

	id, ok := filter[0].Value.(counterID)
	if !ok || id.ClientKey != clientKey || id.WindowStart != bson.DateTime(windowStart.UnixMilli()) {
		t.Fatalf("expected counter ID for %q at %s, got %#v", clientKey, windowStart, filter[0].Value)
	}
}

func assertCounterUpdate(t *testing.T, value any, expiresAt time.Time) {
	t.Helper()

	update, ok := value.(bson.D)
	if !ok || len(update) != 2 || update[0].Key != "$inc" || update[1].Key != "$setOnInsert" {
		t.Fatalf("expected increment and insert update, got %#v", value)
	}

	increment, ok := update[0].Value.(bson.D)
	if !ok || len(increment) != 1 || increment[0].Key != "count" || increment[0].Value != 1 {
		t.Fatalf("expected count increment, got %#v", update[0].Value)
	}

	setOnInsert, ok := update[1].Value.(bson.D)
	if !ok || len(setOnInsert) != 1 || setOnInsert[0].Key != "expires_at" || setOnInsert[0].Value != expiresAt {
		t.Fatalf("expected expiration %s, got %#v", expiresAt, update[1].Value)
	}
}

func findOneAndUpdateOptions(t *testing.T, values []options.Lister[options.FindOneAndUpdateOptions]) options.FindOneAndUpdateOptions {
	t.Helper()

	var result options.FindOneAndUpdateOptions
	for _, value := range values {
		for _, apply := range value.List() {
			if err := apply(&result); err != nil {
				t.Fatalf("expected valid update option: %v", err)
			}
		}
	}

	return result
}

func validIncrementParams() ratelimit.IncrementParams {
	windowStart := time.Date(2026, 7, 13, 5, 0, 0, 0, time.UTC)
	return ratelimit.IncrementParams{
		ClientKey:   "203.0.113.10",
		WindowStart: windowStart,
		ExpiresAt:   windowStart.Add(time.Minute),
	}
}

type fakeCollection struct {
	called  bool
	filter  any
	update  any
	options []options.Lister[options.FindOneAndUpdateOptions]
	result  *mongo.SingleResult
}

func (c *fakeCollection) FindOneAndUpdate(_ context.Context, filter any, update any, options ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	c.called = true
	c.filter = filter
	c.update = update
	c.options = options

	return c.result
}
