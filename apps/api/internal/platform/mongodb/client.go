package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tapadar13/url-shortener/apps/api/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

type Client struct {
	client         *mongo.Client
	database       *mongo.Database
	urlsCollection *mongo.Collection
}

func Connect(ctx context.Context, cfg config.MongoDBConfig, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		return nil, errors.New("MongoDB connection timeout must be greater than zero")
	}

	driverClient, err := mongo.Connect(options.Client().
		ApplyURI(cfg.URI).
		SetServerSelectionTimeout(timeout))
	if err != nil {
		return nil, fmt.Errorf("create MongoDB client: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := driverClient.Ping(pingCtx, readpref.Primary()); err != nil {
		_ = driverClient.Disconnect(context.Background())
		return nil, fmt.Errorf("ping MongoDB: %w", err)
	}

	database := driverClient.Database(cfg.Database)

	return &Client{
		client:         driverClient,
		database:       database,
		urlsCollection: database.Collection(cfg.URLsCollection),
	}, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}

	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("disconnect MongoDB: %w", err)
	}

	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return errors.New("MongoDB client is required")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if err := c.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping MongoDB: %w", err)
	}

	return nil
}

func (c *Client) Database() *mongo.Database {
	if c == nil {
		return nil
	}

	return c.database
}

func (c *Client) URLsCollection() *mongo.Collection {
	if c == nil {
		return nil
	}

	return c.urlsCollection
}
