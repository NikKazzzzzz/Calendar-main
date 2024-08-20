package mongodb

import (
	"context"
	"errors"
	"github.com/NikKazzzzzz/Calendar-main/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"time"
)

type Storage struct {
	client     *mongo.Client
	collection *mongo.Collection
	log        *slog.Logger
}

func New(mongoURI string, dbName, username, password string, log *slog.Logger) (*Storage, error) {
	clientOptions := options.Client().ApplyURI(mongoURI)

	if username != "" && password != "" {
		clientOptions.SetAuth(options.Credential{
			Username: username,
			Password: password,
		})
	}
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection("events")
	storage := &Storage{client: client, collection: collection, log: log}

	return storage, nil
}

func (s *Storage) AddEvent(event *models.Event) error {
	_, err := s.collection.InsertOne(context.Background(), event)
	return err
}

func (s *Storage) UpdateEvent(event *models.Event) error {
	filter := bson.M{"_id": event.ID}
	update := bson.M{"$set": event}
	_, err := s.collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (s *Storage) DeleteEvent(id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := s.collection.DeleteOne(context.Background(), filter)
	return err
}

func (s *Storage) GetEventByID(id primitive.ObjectID) (*models.Event, error) {
	filter := bson.M{"_id": id}
	event := &models.Event{}
	err := s.collection.FindOne(context.Background(), filter).Decode(event)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return event, nil
}

func (s *Storage) GetAllEvent() ([]*models.Event, error) {
	cursor, err := s.collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var events []*models.Event
	for cursor.Next(context.Background()) {
		event := &models.Event{}
		err := cursor.Decode(event)
		if err != nil {
			return nil, err
		}
		events = append(events, event)

		if err := cursor.Err(); err != nil {
			return nil, err
		}
	}

	return events, nil
}

func (s *Storage) IsTimeSlotTaken(startTime time.Time, endTime time.Time) (bool, error) {
	filter := bson.M{
		"$or": []bson.M{
			{"start_time": bson.M{"$lt": startTime}, "end_time": bson.M{"$gt": endTime}},
			{"start_time": bson.M{"$gte": startTime}, "end_time": bson.M{"$lte": endTime}},
		},
	}
	count, err := s.collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
