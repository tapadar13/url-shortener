package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/analytics"
	"github.com/tapadar13/url-shortener/apps/api/internal/shortcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRepositoryRecordClickAtomicallyUpdatesDailyAggregate(t *testing.T) {
	t.Parallel()

	clickedAt := time.Date(2026, 7, 14, 2, 15, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	collection := &fakeCollection{result: &mongo.UpdateResult{MatchedCount: 1, ModifiedCount: 1}}
	repository := newRepository(collection)

	err := repository.RecordClick(context.Background(), analytics.Click{
		ShortCode: " AbC123 ",
		ClickedAt: clickedAt,
	})
	if err != nil {
		t.Fatalf("expected click recording to succeed: %v", err)
	}

	assertClickFilter(t, collection.filter, "AbC123", time.Date(2026, 7, 13, 0, 0, 0, 0, time.UTC))
	assertClickUpdate(t, collection.update, clickedAt.UTC())

	updateOptions := updateOneOptions(t, collection.options)
	if updateOptions.Upsert == nil || !*updateOptions.Upsert {
		t.Fatalf("expected upsert option, got %#v", updateOptions.Upsert)
	}
}

func TestRepositoryRecordClickRejectsNilCollection(t *testing.T) {
	t.Parallel()

	err := New(nil).RecordClick(context.Background(), validClick())
	if err == nil || !strings.Contains(err.Error(), "collection") {
		t.Fatalf("expected collection error, got %v", err)
	}
}

func TestRepositoryRecordClickValidatesClickBeforeQuerying(t *testing.T) {
	t.Parallel()

	collection := &fakeCollection{}
	err := newRepository(collection).RecordClick(context.Background(), analytics.Click{ShortCode: "invalid-code"})
	if !errors.Is(err, shortcode.ErrInvalidChars) {
		t.Fatalf("expected short code error, got %v", err)
	}

	if collection.called {
		t.Fatal("expected collection not to be called")
	}
}

func TestRepositoryRecordClickWrapsDatabaseError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("database unavailable")
	repository := newRepository(&fakeCollection{err: expectedErr})

	err := repository.RecordClick(context.Background(), validClick())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected database error, got %v", err)
	}
}

func TestRepositoryRecordClickRejectsMissingUpdateResult(t *testing.T) {
	t.Parallel()

	err := newRepository(&fakeCollection{}).RecordClick(context.Background(), validClick())
	if err == nil || !strings.Contains(err.Error(), "missing update result") {
		t.Fatalf("expected missing result error, got %v", err)
	}
}

func TestRepositoryRecordClickRetriesConcurrentUpsertCollision(t *testing.T) {
	t.Parallel()

	duplicateErr := mongo.WriteException{
		WriteErrors: mongo.WriteErrors{{Code: 11000, Message: "duplicate key error"}},
	}
	collection := &fakeCollection{
		errors:  []error{duplicateErr, nil},
		results: []*mongo.UpdateResult{nil, {MatchedCount: 1, ModifiedCount: 1}},
	}

	if err := newRepository(collection).RecordClick(context.Background(), validClick()); err != nil {
		t.Fatalf("expected duplicate upsert to be retried: %v", err)
	}

	if collection.calls != 2 {
		t.Fatalf("expected two update attempts, got %d", collection.calls)
	}

	firstOptions := updateOneOptions(t, collection.optionSets[0])
	if firstOptions.Upsert == nil || !*firstOptions.Upsert {
		t.Fatalf("expected first operation to upsert, got %#v", firstOptions.Upsert)
	}

	secondOptions := updateOneOptions(t, collection.optionSets[1])
	if secondOptions.Upsert == nil || *secondOptions.Upsert {
		t.Fatalf("expected retry not to upsert, got %#v", secondOptions.Upsert)
	}
}

func assertClickFilter(t *testing.T, value any, shortCode string, dayStart time.Time) {
	t.Helper()

	filter, ok := value.(bson.D)
	if !ok || len(filter) != 2 {
		t.Fatalf("expected short code and day filter, got %#v", value)
	}

	if filter[0].Key != "short_code" || filter[0].Value != shortCode {
		t.Fatalf("expected short code %q, got %#v", shortCode, filter[0])
	}

	if filter[1].Key != "day_start" || filter[1].Value != dayStart {
		t.Fatalf("expected day start %s, got %#v", dayStart, filter[1])
	}
}

func assertClickUpdate(t *testing.T, value any, clickedAt time.Time) {
	t.Helper()

	update, ok := value.(bson.D)
	if !ok || len(update) != 3 {
		t.Fatalf("expected increment, minimum, and maximum updates, got %#v", value)
	}

	assertSingleFieldUpdate(t, update[0], "$inc", "click_count", 1)
	assertSingleFieldUpdate(t, update[1], "$min", "first_clicked_at", clickedAt)
	assertSingleFieldUpdate(t, update[2], "$max", "last_clicked_at", clickedAt)
}

func assertSingleFieldUpdate(t *testing.T, element bson.E, operator string, field string, expected any) {
	t.Helper()

	if element.Key != operator {
		t.Fatalf("expected %s update, got %#v", operator, element)
	}

	fields, ok := element.Value.(bson.D)
	if !ok || len(fields) != 1 || fields[0].Key != field || fields[0].Value != expected {
		t.Fatalf("expected %s=%v, got %#v", field, expected, element.Value)
	}
}

func updateOneOptions(t *testing.T, values []options.Lister[options.UpdateOneOptions]) options.UpdateOneOptions {
	t.Helper()

	var result options.UpdateOneOptions
	for _, value := range values {
		for _, apply := range value.List() {
			if err := apply(&result); err != nil {
				t.Fatalf("expected valid update option: %v", err)
			}
		}
	}

	return result
}

func validClick() analytics.Click {
	return analytics.Click{
		ShortCode: "AbC123",
		ClickedAt: time.Date(2026, 7, 14, 5, 0, 0, 0, time.UTC),
	}
}

type fakeCollection struct {
	called     bool
	calls      int
	filter     any
	update     any
	options    []options.Lister[options.UpdateOneOptions]
	optionSets [][]options.Lister[options.UpdateOneOptions]
	result     *mongo.UpdateResult
	results    []*mongo.UpdateResult
	err        error
	errors     []error
}

func (c *fakeCollection) UpdateOne(_ context.Context, filter any, update any, options ...options.Lister[options.UpdateOneOptions]) (*mongo.UpdateResult, error) {
	c.called = true
	callIndex := c.calls
	c.calls++
	c.filter = filter
	c.update = update
	c.options = options
	c.optionSets = append(c.optionSets, options)

	if callIndex < len(c.results) || callIndex < len(c.errors) {
		var result *mongo.UpdateResult
		if callIndex < len(c.results) {
			result = c.results[callIndex]
		}

		var err error
		if callIndex < len(c.errors) {
			err = c.errors[callIndex]
		}

		return result, err
	}

	return c.result, c.err
}
