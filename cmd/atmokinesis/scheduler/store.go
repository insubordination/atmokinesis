package scheduler

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

type Store interface {
	UpdateEntries(c context.Context, entries []*Entry) (err error)
	AddEntries(c context.Context, entries []*Entry) (err error)
	UpdateInMemoryEntriesFromStorage(c context.Context, entries []*Entry) (err error)
	Close(c context.Context) error
}

const (
	database   = "atmokinesis"
	collection = "entries"
)

func NewMongoStore(ctx context.Context, uri string) (Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return &mongoStore{client: client, entryCollection: client.Database(database).Collection(collection)}, nil
}

type mongoStore struct {
	client          *mongo.Client
	entryCollection *mongo.Collection
}

func (s mongoStore) UpdateEntries(c context.Context, entries []*Entry) (err error) {
	for _, e := range entries {
		ctx, cancel := context.WithTimeout(c, 60*time.Second)
		data, marshalErr := e.MarshalToBSON()
		if marshalErr != nil {
			errors.As(fmt.Errorf("update to entry had an issue: %w", marshalErr), &err)
			cancel()
			continue
		}
		res, updateErr := s.entryCollection.UpdateOne(ctx, bson.D{{"_id", e.Task.TaskID()}}, bson.M{"$addToSet": data})
		if updateErr != nil {
			errors.As(fmt.Errorf("update to entry had an issue: %w", updateErr), &err)
			cancel()
			continue
		}
		if res.MatchedCount == 0 {
			if addErr := s.AddEntries(c, []*Entry{e}); addErr != nil {
				errors.As(fmt.Errorf("adding entry failed: %w", addErr), &err)
			}
		}
		cancel()
	}
	return err
}

func (s mongoStore) AddEntries(c context.Context, entries []*Entry) (err error) {
	for _, e := range entries {
		ctx, cancel := context.WithTimeout(c, 60*time.Second)
		data, marshalErr := e.MarshalToBSON()
		if marshalErr != nil {
			errors.As(fmt.Errorf("adding entry had an issue: %w", marshalErr), &err)
			cancel()
			continue
		}
		data["_id"] = e.Task.TaskID()
		_, updateErr := s.entryCollection.InsertOne(ctx, data)
		if updateErr != nil {
			errors.As(fmt.Errorf("adding entry had an issue: %w", updateErr), &err)
		}
		cancel()
	}
	return err
}

func (s mongoStore) UpdateInMemoryEntriesFromStorage(c context.Context, entries []*Entry) (err error) {
	for _, e := range entries {
		ctx, cancel := context.WithTimeout(c, 60*time.Second)
		body, bodyErr := e.MarshalToBSON()
		if bodyErr != nil {
			errors.As(fmt.Errorf("loading entry from store failed on decoding: %w", bodyErr), &err)
			cancel()
			continue
		}

		res := s.entryCollection.FindOne(ctx, bson.D{{"_id", e.Task.TaskID()}})
		if decErr := res.Decode(&body); decErr != nil {
			errors.As(fmt.Errorf("loading entry from store failed on decoding: %w", decErr), &err)
			cancel()
			continue
		}

		if unmarshalErr := UnmarshalBSON(body, e); err != nil {
			errors.As(fmt.Errorf("loading entry from store failed on updating entry: %w", unmarshalErr), &err)
			cancel()
			continue
		}
		cancel()
	}
	return err
}

func (s mongoStore) Close(c context.Context) error {
	ctx, cancel := context.WithTimeout(c, 160*time.Second)
	defer cancel()
	return s.client.Disconnect(ctx)
}
