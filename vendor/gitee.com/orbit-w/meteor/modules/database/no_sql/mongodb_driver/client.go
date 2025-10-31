package mongodbdriver

import (
	"context"

	"gitee.com/orbit-w/meteor/modules/mlog"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoClient provides access to MongoDB for persistent storage
type VirtualMongoClient struct {
	client *mongo.Client
	config MongoDBConfig
	logger *mlog.Logger
}

// NewMongoClient creates a new MongoDB client
func NewMongoClient(cfg MongoDBConfig) (*VirtualMongoClient, error) {
	clientOptions := options.Client().
		ApplyURI(cfg.URI).
		SetMaxPoolSize(cfg.MaxPoolSize).
		SetMinPoolSize(cfg.MinPoolSize).
		SetMaxConnIdleTime(cfg.MaxConnIdleTime).
		SetMaxConnecting(cfg.MaxConnecting).
		SetRetryWrites(cfg.RetryWrites).
		SetRetryReads(cfg.RetryReads).
		SetConnectTimeout(cfg.ConnectTimeout)

	cli, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	virtualClient := &VirtualMongoClient{
		client: cli,
		config: cfg,
		logger: mlog.WithPrefix("MongoClient"),
	}

	err = virtualClient.Ping()
	if err != nil {
		_ = virtualClient.Disconnect(context.TODO())
		return nil, err
	}

	return virtualClient, nil
}

func (c *VirtualMongoClient) Client() *mongo.Client {
	return c.client
}

// Database returns a handle for a database with the given name configured with the given DatabaseOptions.
func (c *VirtualMongoClient) Database(name string) *mongo.Database {
	return c.client.Database(name)
}

func (c *VirtualMongoClient) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.PingTimeout)
	defer cancel()
	return c.client.Ping(ctx, readpref.Primary())
}

func (c *VirtualMongoClient) Disconnect(ctx context.Context) error {
	if ctx == nil {
		var cancel func()
		ctx, cancel = context.WithTimeout(context.Background(), c.config.DisconnectTimeout)
		defer cancel()
	}

	return c.client.Disconnect(ctx)
}
