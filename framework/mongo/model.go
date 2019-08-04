package mongo

import (
	"context"
	"fmt"

	"github.com/kukinsula/boxy/entity"
	"github.com/kukinsula/boxy/entity/log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type model struct {
	collection *mongo.Collection
	params     modelParams
}

type modelParams struct {
	Context  context.Context
	Database *Database
	Name     string
	Logger   log.Logger
	Indexes  []indexParams
}

type indexParams struct {
	Name       string
	Unique     bool
	Value      int32
	Background bool
	Sparse     bool
}

func newModel(params modelParams) (*model, error) {
	model := &model{
		collection: params.Database.database.Collection(params.Name),
		params:     params,
	}

	uuid := entity.NewUUID()

	for _, index := range params.Indexes {
		err := model.createindex(uuid, params.Context, mongo.IndexModel{
			Keys: bsonx.Doc{{Key: index.Name, Value: bsonx.Int32(index.Value)}},
			Options: options.Index().
				SetName(index.Name).
				SetUnique(index.Unique).
				SetBackground(index.Background).
				SetSparse(index.Sparse).
				SetStorageEngine(bsonx.Doc{{
					"wiredTiger", bsonx.Document(bsonx.Doc{{
						"configString", bsonx.String("block_compressor=zlib"),
					}})},
				}),
		})

		if err != nil {
			return nil, err
		}
	}

	return model, nil
}

func (model *model) createindex(
	uuid string,
	ctx context.Context,
	index mongo.IndexModel,
	opts ...*options.CreateIndexesOptions) error {

	_, err := model.collection.Indexes().CreateOne(ctx, index, opts...)

	model.params.Logger(uuid, log.DEBUG,
		fmt.Sprintf("%s.EnsureIndex", model.params.Name),
		map[string]interface{}{
			"name":   *index.Options.Name,
			"unique": *index.Options.Unique,
		})

	return err
}

func (model *model) InsertOne(uuid string, ctx context.Context, data interface{}) error {
	_, err := model.collection.InsertOne(ctx, data)

	model.params.Logger(uuid, log.DEBUG,
		fmt.Sprintf("%s.InsertOne", model.params.Name),
		map[string]interface{}{"data": data, "error": err})

	return err
}

func (model *model) FindOne(
	uuid string,
	ctx context.Context,
	conditions map[string]interface{},
	projection map[string]interface{},
	result interface{}) error {

	opts := &options.FindOneOptions{Projection: projection}
	err := model.collection.FindOne(ctx, conditions, opts).Decode(result)

	model.params.Logger(uuid, log.DEBUG,
		fmt.Sprintf("%s.FindOne", model.params.Name),
		map[string]interface{}{
			"conditions": conditions,
			"projection": projection,
			"result":     result,
			"error":      err,
		})

	return err
}

func (model *model) UpdateOne(
	uuid string,
	ctx context.Context,
	conditions map[string]interface{},
	update map[string]interface{},
	opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {

	result, err := model.collection.UpdateOne(ctx, conditions, update, opts...)
	if err != nil {
		return nil, err
	}

	if result.MatchedCount != 1 || result.ModifiedCount != 1 {
		return nil, fmt.Errorf("Model.UpdateOne failed: wrong number of Matched or Modified counts")
	}

	model.params.Logger(uuid, log.DEBUG,
		fmt.Sprintf("%s.UpdateOne", model.params.Name),
		map[string]interface{}{
			"conditions": conditions,
			"update":     update,
			"result":     result,
		})

	return result, nil
}

func (model *model) DeleteOne(
	uuid string,
	ctx context.Context,
	conditions map[string]interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	result, err := model.collection.DeleteOne(ctx, conditions, opts...)
	if err != nil {
		return nil, err
	}

	if result.DeletedCount != 1 {
		return nil, fmt.Errorf("Model.DeleteOne failed: wrong number of Deleted counts")
	}

	model.params.Logger(uuid, log.DEBUG,
		fmt.Sprintf("%s.DeleteOne", model.params.Name),
		map[string]interface{}{"conditions": conditions})

	return result, nil
}

func (model *model) DeleteMany(
	uuid string,
	ctx context.Context,
	conditions map[string]interface{},
	opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {

	result, err := model.collection.DeleteMany(ctx, conditions, opts...)
	if err != nil {
		return nil, err
	}

	model.params.Logger(uuid, log.DEBUG,
		fmt.Sprintf("%s.DeleteMany", model.params.Name),
		map[string]interface{}{"conditions": conditions})

	return result, nil
}

func (model *model) RemoveAllDocuments(uuid string, ctx context.Context) (*mongo.DeleteResult, error) {
	result, err := model.collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	return result, nil
}
