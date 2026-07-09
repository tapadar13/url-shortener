package mongodb

import (
	"context"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func TestURLIndexModels(t *testing.T) {
	t.Parallel()

	models := URLIndexModels()
	if len(models) != 2 {
		t.Fatalf("expected 2 index models, got %d", len(models))
	}

	assertIndexKeys(t, models[0].Keys, bson.D{{Key: "short_code", Value: 1}})
	assertIndexOptions(t, models[0].Options, ShortCodeIndexName, true)

	assertIndexKeys(t, models[1].Keys, bson.D{{Key: "created_at", Value: -1}})
	assertIndexOptions(t, models[1].Options, CreatedAtIndexName, false)
}

func TestEnsureIndexesRejectsNilClient(t *testing.T) {
	t.Parallel()

	err := EnsureIndexes(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "client") {
		t.Fatalf("expected client error, got %q", err.Error())
	}
}

func TestEnsureIndexesRejectsNilCollection(t *testing.T) {
	t.Parallel()

	err := EnsureIndexes(context.Background(), &Client{})
	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "collection") {
		t.Fatalf("expected collection error, got %q", err.Error())
	}
}

func assertIndexKeys(t *testing.T, actual any, expected bson.D) {
	t.Helper()

	actualKeys, ok := actual.(bson.D)
	if !ok {
		t.Fatalf("expected bson.D keys, got %T", actual)
	}

	if len(actualKeys) != len(expected) {
		t.Fatalf("expected %d keys, got %d", len(expected), len(actualKeys))
	}

	for i := range expected {
		if actualKeys[i] != expected[i] {
			t.Fatalf("expected key %d to be %+v, got %+v", i, expected[i], actualKeys[i])
		}
	}
}

func assertIndexOptions(t *testing.T, builder *options.IndexOptionsBuilder, expectedName string, expectedUnique bool) {
	t.Helper()

	if builder == nil {
		t.Fatal("expected index options")
	}

	var opts options.IndexOptions
	for _, apply := range builder.List() {
		if err := apply(&opts); err != nil {
			t.Fatalf("expected option to apply cleanly: %v", err)
		}
	}

	if opts.Name == nil || *opts.Name != expectedName {
		t.Fatalf("expected index name %q, got %+v", expectedName, opts.Name)
	}

	if expectedUnique {
		if opts.Unique == nil || !*opts.Unique {
			t.Fatalf("expected unique index, got %+v", opts.Unique)
		}

		return
	}

	if opts.Unique != nil && *opts.Unique {
		t.Fatalf("did not expect unique index")
	}
}
