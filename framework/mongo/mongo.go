package mongo

import (
	"context"
	"time"

	"github.com/kukinsula/boxy/entity/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Database struct {
	client   *mongo.Client
	database *mongo.Database
	User     *UserModel
	params   NewDatabaseParams
}

type NewDatabaseParams struct {
	Context  context.Context
	URI      string
	Database string
	Timeout  time.Duration
	Logger   log.Logger
}

func NewDatabase(params NewDatabaseParams) (*Database, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(params.URI))
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(params.Context, params.Timeout)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	return &Database{
		client:   client,
		database: client.Database(params.Database),
		params:   params,
	}, nil
}

func (database *Database) Init(ctx context.Context) error {
	user, err := NewUserModel(ctx, database, database.params.Logger)
	if err != nil {
		return err
	}

	database.User = user

	return nil
}

func (database *Database) Drop(ctx context.Context) error {
	return database.database.Drop(ctx)
}
