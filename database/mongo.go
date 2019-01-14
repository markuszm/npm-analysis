package database

import (
	"context"
	"github.com/markuszm/npm-analysis/evolution"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/x/bsonx"
	"time"
)

type MongoDB struct {
	url, database, collection string
	Client                    *mongo.Client
	ActiveCollection          *mongo.Collection
}

func NewMongoDB(url, database, collection string) *MongoDB {
	return &MongoDB{url: url, database: database, collection: collection}
}

func (m *MongoDB) Connect() error {
	mongodb, err := mongo.NewClient(m.url)
	if err != nil {
		return err
	}
	m.Client = mongodb
	err = mongodb.Connect(context.Background())
	m.ActiveCollection = mongodb.Database(m.database).Collection(m.collection)
	return err
}

func (m *MongoDB) Disconnect() {
	m.Client.Disconnect(context.Background())
}

func (m *MongoDB) EnsureSingleIndex(key string) (string, error) {
	return m.ActiveCollection.Indexes().CreateOne(context.Background(),
		mongo.IndexModel{
			Keys: bson.D{
				{key, 1}}})
}

func (m *MongoDB) FindOneSimple(key, value string) (string, error) {
	result := m.ActiveCollection.FindOne(context.Background(), bson.D{
		{key, value},
	})

	val := Document{}

	err := result.Decode(&val)
	if err != nil {
		return "", err
	}
	element := val.Value
	if element == "" {
		return "", err
	}
	return element, nil
}

func (m *MongoDB) FindAllSimple(key, value string) ([]string, error) {
	var result []string
	cursor, err := m.ActiveCollection.Find(context.Background(), bson.D{
		{key, value},
	})
	if err != nil {
		return result, nil
	}

	for cursor.Next(context.Background()) {
		val, err := m.DecodeValue(cursor)
		if err != nil {
			return result, err
		}
		result = append(result, val.Value)
	}

	return result, nil
}

func (m *MongoDB) FindPackageDataInTimeline(pkg string, time time.Time) (evolution.PackageData, error) {
	result := m.ActiveCollection.FindOne(context.Background(), bson.D{
		{"key", pkg},
		{"timeline.time", time.String()},
	}, options.FindOne().SetProjection(bson.D{{"_id", 0}, {"timeline.$", 1}}))

	var val bsonx.Doc

	err := result.Decode(&val)
	if err != nil {
		return evolution.PackageData{}, err
	}

	timeline := val.Lookup("timeline")
	if timeline.IsZero() {
		return evolution.PackageData{}, nil
	}
	pkgDataBson := timeline.Array()[0].Document().Lookup("packageData")

	maintainersBson := pkgDataBson.Document().Lookup("maintainers")
	dependenciesBson := pkgDataBson.Document().Lookup("dependencies")

	maintainers := make([]string, 0)
	dependencies := make([]string, 0)

	if !maintainersBson.IsZero() {
		for _, m := range maintainersBson.Array() {
			maintainers = append(maintainers, m.String())
		}
	}

	if !dependenciesBson.IsZero() {
		for _, d := range dependenciesBson.Array() {
			dependencies = append(dependencies, d.String())
		}
	}

	packageData := evolution.PackageData{
		Version:      pkgDataBson.Document().Lookup("version").String(),
		Maintainers:  maintainers,
		Dependencies: dependencies,
	}

	return packageData, nil
}

func (m *MongoDB) FindAll() ([]Document, error) {
	var result []Document
	cursor, err := m.ActiveCollection.Find(context.Background(), bson.D{})
	if err != nil {
		return result, err
	}
	for cursor.Next(context.Background()) {
		val, err := m.DecodeValue(cursor)
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}
	return result, nil
}

func (m *MongoDB) DecodeValue(cur mongo.Cursor) (Document, error) {
	val := Document{}

	err := cur.Decode(&val)
	if err != nil {
		return Document{}, err
	}
	return val, nil
}

func (m *MongoDB) InsertOneSimple(key, value string) error {
	_, err := m.ActiveCollection.InsertOne(context.Background(), bson.D{
		{"key", key},
		{"value", value},
	})
	return err
}

func (m *MongoDB) InsertPackageTimeline(pkg string, pkgTimeline map[time.Time]evolution.PackageData) error {
	var elements []bson.M
	for t, d := range pkgTimeline {
		elements = append(elements, bson.M{"time": t.String(), "packageData": d})
	}

	pkgDocument := bson.D{
		{Key: "key", Value: pkg},
		{Key: "timeline", Value: elements},
	}

	_, err := m.ActiveCollection.InsertOne(context.Background(), pkgDocument)
	return err
}

func (m *MongoDB) RemoveWithKey(key string) error {
	_, err := m.ActiveCollection.DeleteOne(context.Background(), bson.D{
		{"key", key}})
	return err
}

type Document struct {
	Key, Value string
}
