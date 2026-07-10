package mongodb

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	urlmodel "github.com/tapadar13/url-shortener/apps/api/internal/url"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestRepositoryCreateInsertsURLDocument(t *testing.T) {
	t.Parallel()

	insertedID := bson.NewObjectID()
	collection := &fakeInsertOneCollection{
		result: &mongo.InsertOneResult{
			InsertedID:   insertedID,
			Acknowledged: true,
		},
	}

	record := newValidURLRecord(t)
	repository := newRepository(collection)

	created, err := repository.Create(context.Background(), record)
	if err != nil {
		t.Fatalf("expected URL to be created: %v", err)
	}

	if collection.insertCount != 1 {
		t.Fatalf("expected one insert, got %d", collection.insertCount)
	}

	insertedDoc, ok := collection.document.(urlDocument)
	if !ok {
		t.Fatalf("expected urlDocument insert, got %T", collection.document)
	}

	if insertedDoc.ShortCode != record.ShortCode {
		t.Fatalf("expected short code %q, got %q", record.ShortCode, insertedDoc.ShortCode)
	}

	if created.ID != insertedID.Hex() {
		t.Fatalf("expected created id %q, got %q", insertedID.Hex(), created.ID)
	}

	if created.LongURL != record.LongURL {
		t.Fatalf("expected URL %q, got %q", record.LongURL, created.LongURL)
	}
}

func TestRepositoryCreateRejectsNilCollection(t *testing.T) {
	t.Parallel()

	_, err := New(nil).Create(context.Background(), newValidURLRecord(t))
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "collection") {
		t.Fatalf("expected collection error, got %q", err.Error())
	}
}

func TestRepositoryCreateValidatesRecordBeforeInsert(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{}
	repository := newRepository(collection)

	_, err := repository.Create(context.Background(), urlmodel.URL{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if collection.insertCount != 0 {
		t.Fatalf("expected no insert, got %d", collection.insertCount)
	}
}

func TestRepositoryCreateMapsDuplicateShortCodeError(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{
		err: mongo.WriteException{
			WriteErrors: mongo.WriteErrors{
				{
					Code:    11000,
					Message: "duplicate key error",
				},
			},
		},
	}
	repository := newRepository(collection)

	_, err := repository.Create(context.Background(), newValidURLRecord(t))
	if !errors.Is(err, urlmodel.ErrDuplicateShortCode) {
		t.Fatalf("expected duplicate short code error, got %v", err)
	}
}

func TestRepositoryCreateWrapsInsertError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("insert failed")
	collection := &fakeInsertOneCollection{err: expectedErr}
	repository := newRepository(collection)

	_, err := repository.Create(context.Background(), newValidURLRecord(t))
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected insert error, got %v", err)
	}
}

func TestRepositoryCreateRejectsUnexpectedInsertedID(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{
		result: &mongo.InsertOneResult{
			InsertedID: "not-object-id",
		},
	}
	repository := newRepository(collection)

	_, err := repository.Create(context.Background(), newValidURLRecord(t))
	if err == nil {
		t.Fatal("expected inserted id error")
	}

	if !strings.Contains(err.Error(), "expected ObjectID") {
		t.Fatalf("expected ObjectID error, got %q", err.Error())
	}
}

func newValidURLRecord(t *testing.T) urlmodel.URL {
	t.Helper()

	record, err := urlmodel.New(urlmodel.NewParams{
		LongURL:   "https://example.com/articles/123",
		ShortCode: "AbC123",
		Now:       time.Date(2026, 7, 9, 7, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("expected valid URL record: %v", err)
	}

	return record
}

type fakeInsertOneCollection struct {
	document    any
	result      *mongo.InsertOneResult
	err         error
	insertCount int
}

func (c *fakeInsertOneCollection) InsertOne(_ context.Context, document any, _ ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	c.insertCount++
	c.document = document

	return c.result, c.err
}
