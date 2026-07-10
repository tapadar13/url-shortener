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

func TestRepositoryFindByShortCodeReturnsURL(t *testing.T) {
	t.Parallel()

	record := newValidURLRecord(t)
	id := bson.NewObjectID()
	collection := &fakeInsertOneCollection{
		findResult: mongo.NewSingleResultFromDocument(urlDocument{
			ID:          id,
			LongURL:     record.LongURL,
			ShortCode:   record.ShortCode,
			AccessCount: 8,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   record.UpdatedAt,
		}, nil, nil),
	}
	repository := newRepository(collection)

	found, err := repository.FindByShortCode(context.Background(), " "+record.ShortCode+" ")
	if err != nil {
		t.Fatalf("expected URL to be found: %v", err)
	}

	if collection.findCount != 1 {
		t.Fatalf("expected one lookup, got %d", collection.findCount)
	}

	filter, ok := collection.filter.(bson.D)
	if !ok {
		t.Fatalf("expected BSON filter, got %T", collection.filter)
	}

	if len(filter) != 1 || filter[0].Key != "short_code" || filter[0].Value != record.ShortCode {
		t.Fatalf("expected short code filter, got %#v", filter)
	}

	if found.ID != id.Hex() || found.LongURL != record.LongURL || found.AccessCount != 8 {
		t.Fatalf("expected found URL record, got %#v", found)
	}
}

func TestRepositoryFindByShortCodeMapsMissingDocument(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{
		findResult: mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil),
	}
	repository := newRepository(collection)

	_, err := repository.FindByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, urlmodel.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRepositoryFindByShortCodeValidatesShortCodeBeforeQuerying(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{}
	repository := newRepository(collection)

	_, err := repository.FindByShortCode(context.Background(), "invalid-code")
	if err == nil {
		t.Fatal("expected validation error")
	}

	if collection.findCount != 0 {
		t.Fatalf("expected no lookup, got %d", collection.findCount)
	}
}

func TestRepositoryFindByShortCodeWrapsQueryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("query failed")
	collection := &fakeInsertOneCollection{
		findResult: mongo.NewSingleResultFromDocument(bson.D{}, expectedErr, nil),
	}
	repository := newRepository(collection)

	_, err := repository.FindByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected query error, got %v", err)
	}
}

func TestRepositoryUpdateLongURLReturnsUpdatedURL(t *testing.T) {
	t.Parallel()

	record := newValidURLRecord(t)
	id := bson.NewObjectID()
	updatedAt := time.Date(2026, 7, 10, 10, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	collection := &fakeInsertOneCollection{
		updateResult: mongo.NewSingleResultFromDocument(urlDocument{
			ID:          id,
			LongURL:     "https://example.com/updated",
			ShortCode:   record.ShortCode,
			AccessCount: 8,
			CreatedAt:   record.CreatedAt,
			UpdatedAt:   updatedAt.UTC(),
		}, nil, nil),
	}
	repository := newRepository(collection)

	updated, err := repository.UpdateLongURL(context.Background(), urlmodel.UpdateLongURLParams{
		ShortCode: " " + record.ShortCode + " ",
		LongURL:   " https://example.com/updated ",
		UpdatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("expected URL update to succeed: %v", err)
	}

	if collection.updateCount != 1 {
		t.Fatalf("expected one update, got %d", collection.updateCount)
	}

	assertShortCodeFilter(t, collection.updateFilter, record.ShortCode)
	assertURLUpdate(t, collection.update, "https://example.com/updated", updatedAt.UTC())

	updateOptions := findOneAndUpdateOptions(t, collection.updateOptions)
	if updateOptions.ReturnDocument == nil || *updateOptions.ReturnDocument != options.After {
		t.Fatalf("expected updated document to be returned, got %#v", updateOptions.ReturnDocument)
	}

	if updated.ID != id.Hex() || updated.LongURL != "https://example.com/updated" || !updated.UpdatedAt.Equal(updatedAt.UTC()) {
		t.Fatalf("expected updated URL record, got %#v", updated)
	}
}

func TestRepositoryUpdateLongURLMapsMissingDocument(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{
		updateResult: mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil),
	}
	repository := newRepository(collection)

	_, err := repository.UpdateLongURL(context.Background(), urlmodel.UpdateLongURLParams{
		ShortCode: "AbC123",
		LongURL:   "https://example.com/updated",
		UpdatedAt: time.Now(),
	})
	if !errors.Is(err, urlmodel.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRepositoryUpdateLongURLValidatesParametersBeforeUpdating(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{}
	repository := newRepository(collection)

	_, err := repository.UpdateLongURL(context.Background(), urlmodel.UpdateLongURLParams{
		ShortCode: "invalid-code",
		LongURL:   "https://example.com/updated",
		UpdatedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if collection.updateCount != 0 {
		t.Fatalf("expected no update, got %d", collection.updateCount)
	}
}

func TestRepositoryUpdateLongURLWrapsQueryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("update failed")
	collection := &fakeInsertOneCollection{
		updateResult: mongo.NewSingleResultFromDocument(bson.D{}, expectedErr, nil),
	}
	repository := newRepository(collection)

	_, err := repository.UpdateLongURL(context.Background(), urlmodel.UpdateLongURLParams{
		ShortCode: "AbC123",
		LongURL:   "https://example.com/updated",
		UpdatedAt: time.Now(),
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected update error, got %v", err)
	}
}

func TestRepositoryDeleteByShortCodeDeletesURL(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{
		deleteResult: &mongo.DeleteResult{
			DeletedCount: 1,
			Acknowledged: true,
		},
	}
	repository := newRepository(collection)

	if err := repository.DeleteByShortCode(context.Background(), " AbC123 "); err != nil {
		t.Fatalf("expected URL deletion to succeed: %v", err)
	}

	if collection.deleteCount != 1 {
		t.Fatalf("expected one delete, got %d", collection.deleteCount)
	}

	assertShortCodeFilter(t, collection.deleteFilter, "AbC123")
}

func TestRepositoryDeleteByShortCodeMapsMissingURL(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{
		deleteResult: &mongo.DeleteResult{Acknowledged: true},
	}
	repository := newRepository(collection)

	err := repository.DeleteByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, urlmodel.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRepositoryDeleteByShortCodeValidatesShortCodeBeforeDeleting(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{}
	repository := newRepository(collection)

	err := repository.DeleteByShortCode(context.Background(), "invalid-code")
	if err == nil {
		t.Fatal("expected validation error")
	}

	if collection.deleteCount != 0 {
		t.Fatalf("expected no delete, got %d", collection.deleteCount)
	}
}

func TestRepositoryDeleteByShortCodeWrapsDeleteError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("delete failed")
	collection := &fakeInsertOneCollection{deleteErr: expectedErr}
	repository := newRepository(collection)

	err := repository.DeleteByShortCode(context.Background(), "AbC123")
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected delete error, got %v", err)
	}
}

func TestRepositoryRecordAccessReturnsURLWithIncrementedCount(t *testing.T) {
	t.Parallel()

	record := newValidURLRecord(t)
	id := bson.NewObjectID()
	accessedAt := time.Date(2026, 7, 10, 10, 0, 0, 0, time.FixedZone("IST", 5*60*60+30*60))
	accessedAtUTC := accessedAt.UTC()
	collection := &fakeInsertOneCollection{
		updateResult: mongo.NewSingleResultFromDocument(urlDocument{
			ID:             id,
			LongURL:        record.LongURL,
			ShortCode:      record.ShortCode,
			AccessCount:    1,
			CreatedAt:      record.CreatedAt,
			UpdatedAt:      record.UpdatedAt,
			LastAccessedAt: &accessedAtUTC,
		}, nil, nil),
	}
	repository := newRepository(collection)

	recorded, err := repository.RecordAccess(context.Background(), urlmodel.RecordAccessParams{
		ShortCode:  " " + record.ShortCode + " ",
		AccessedAt: accessedAt,
	})
	if err != nil {
		t.Fatalf("expected access to be recorded: %v", err)
	}

	if collection.updateCount != 1 {
		t.Fatalf("expected one access update, got %d", collection.updateCount)
	}

	assertShortCodeFilter(t, collection.updateFilter, record.ShortCode)
	assertAccessUpdate(t, collection.update, accessedAtUTC)

	updateOptions := findOneAndUpdateOptions(t, collection.updateOptions)
	if updateOptions.ReturnDocument == nil || *updateOptions.ReturnDocument != options.After {
		t.Fatalf("expected updated document to be returned, got %#v", updateOptions.ReturnDocument)
	}

	if recorded.ID != id.Hex() || recorded.AccessCount != 1 || recorded.LastAccessedAt == nil || !recorded.LastAccessedAt.Equal(accessedAtUTC) {
		t.Fatalf("expected recorded access result, got %#v", recorded)
	}
}

func TestRepositoryRecordAccessMapsMissingURL(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{
		updateResult: mongo.NewSingleResultFromDocument(bson.D{}, mongo.ErrNoDocuments, nil),
	}
	repository := newRepository(collection)

	_, err := repository.RecordAccess(context.Background(), urlmodel.RecordAccessParams{
		ShortCode:  "AbC123",
		AccessedAt: time.Now(),
	})
	if !errors.Is(err, urlmodel.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestRepositoryRecordAccessValidatesParametersBeforeUpdating(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{}
	repository := newRepository(collection)

	_, err := repository.RecordAccess(context.Background(), urlmodel.RecordAccessParams{
		ShortCode:  "invalid-code",
		AccessedAt: time.Now(),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}

	if collection.updateCount != 0 {
		t.Fatalf("expected no access update, got %d", collection.updateCount)
	}
}

func TestRepositoryRecordAccessRequiresTimestamp(t *testing.T) {
	t.Parallel()

	collection := &fakeInsertOneCollection{}
	repository := newRepository(collection)

	_, err := repository.RecordAccess(context.Background(), urlmodel.RecordAccessParams{
		ShortCode: "AbC123",
	})
	if !errors.Is(err, urlmodel.ErrTimestampRequired) {
		t.Fatalf("expected timestamp error, got %v", err)
	}

	if collection.updateCount != 0 {
		t.Fatalf("expected no access update, got %d", collection.updateCount)
	}
}

func TestRepositoryRecordAccessWrapsQueryError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("update failed")
	collection := &fakeInsertOneCollection{
		updateResult: mongo.NewSingleResultFromDocument(bson.D{}, expectedErr, nil),
	}
	repository := newRepository(collection)

	_, err := repository.RecordAccess(context.Background(), urlmodel.RecordAccessParams{
		ShortCode:  "AbC123",
		AccessedAt: time.Now(),
	})
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected access update error, got %v", err)
	}
}

func assertShortCodeFilter(t *testing.T, value any, shortCode string) {
	t.Helper()

	filter, ok := value.(bson.D)
	if !ok {
		t.Fatalf("expected BSON filter, got %T", value)
	}

	if len(filter) != 1 || filter[0].Key != "short_code" || filter[0].Value != shortCode {
		t.Fatalf("expected short code filter, got %#v", filter)
	}
}

func assertURLUpdate(t *testing.T, value any, longURL string, updatedAt time.Time) {
	t.Helper()

	update, ok := value.(bson.D)
	if !ok {
		t.Fatalf("expected BSON update, got %T", value)
	}

	if len(update) != 1 || update[0].Key != "$set" {
		t.Fatalf("expected $set update, got %#v", update)
	}

	fields, ok := update[0].Value.(bson.D)
	if !ok {
		t.Fatalf("expected BSON $set fields, got %T", update[0].Value)
	}

	if len(fields) != 2 || fields[0].Key != "url" || fields[0].Value != longURL || fields[1].Key != "updated_at" || fields[1].Value != updatedAt {
		t.Fatalf("expected URL and updated timestamp fields, got %#v", fields)
	}
}

func assertAccessUpdate(t *testing.T, value any, accessedAt time.Time) {
	t.Helper()

	update, ok := value.(bson.D)
	if !ok {
		t.Fatalf("expected BSON update, got %T", value)
	}

	if len(update) != 2 || update[0].Key != "$inc" || update[1].Key != "$set" {
		t.Fatalf("expected $inc and $set update, got %#v", update)
	}

	increment, ok := update[0].Value.(bson.D)
	if !ok || len(increment) != 1 || increment[0].Key != "access_count" || increment[0].Value != 1 {
		t.Fatalf("expected access count increment, got %#v", update[0].Value)
	}

	fields, ok := update[1].Value.(bson.D)
	if !ok || len(fields) != 1 || fields[0].Key != "last_accessed_at" || fields[0].Value != accessedAt {
		t.Fatalf("expected access timestamp update, got %#v", update[1].Value)
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
	document      any
	result        *mongo.InsertOneResult
	err           error
	insertCount   int
	filter        any
	findResult    *mongo.SingleResult
	findCount     int
	updateFilter  any
	update        any
	updateOptions []options.Lister[options.FindOneAndUpdateOptions]
	updateResult  *mongo.SingleResult
	updateCount   int
	deleteFilter  any
	deleteResult  *mongo.DeleteResult
	deleteErr     error
	deleteCount   int
}

func (c *fakeInsertOneCollection) InsertOne(_ context.Context, document any, _ ...options.Lister[options.InsertOneOptions]) (*mongo.InsertOneResult, error) {
	c.insertCount++
	c.document = document

	return c.result, c.err
}

func (c *fakeInsertOneCollection) FindOne(_ context.Context, filter any, _ ...options.Lister[options.FindOneOptions]) *mongo.SingleResult {
	c.findCount++
	c.filter = filter

	return c.findResult
}

func (c *fakeInsertOneCollection) FindOneAndUpdate(_ context.Context, filter any, update any, options ...options.Lister[options.FindOneAndUpdateOptions]) *mongo.SingleResult {
	c.updateCount++
	c.updateFilter = filter
	c.update = update
	c.updateOptions = options

	return c.updateResult
}

func (c *fakeInsertOneCollection) DeleteOne(_ context.Context, filter any, _ ...options.Lister[options.DeleteOneOptions]) (*mongo.DeleteResult, error) {
	c.deleteCount++
	c.deleteFilter = filter

	return c.deleteResult, c.deleteErr
}
