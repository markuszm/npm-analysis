package evolution

import (
	"context"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

type MongoDB struct {
	url, database, collection string
	client                    *mongo.Client
	activeCollection          *mongo.Collection
}

func NewMongoDB(url, database, collection string) *MongoDB {
	return &MongoDB{url: url, database: database, collection: collection}
}

func (m *MongoDB) Connect() error {
	mongodb, err := mongo.NewClient(m.url)
	if err != nil {
		return err
	}
	m.client = mongodb
	err = mongodb.Connect(context.Background())
	m.activeCollection = mongodb.Database(m.database).Collection(m.collection)
	return err
}

func (m *MongoDB) Disconnect() {
	m.client.Disconnect(context.Background())
}

func (m *MongoDB) EnsureSingleIndex(key string) (string, error) {
	return m.activeCollection.Indexes().CreateOne(context.Background(),
		mongo.IndexModel{
			Keys: bson.NewDocument(
				bson.EC.Int32(key, 1))})
}

func (m *MongoDB) FindOneSimple(key, value string) (string, error) {
	result := m.activeCollection.FindOne(context.Background(), bson.NewDocument(
		bson.EC.String(key, value),
	))

	val := bson.NewDocument()

	err := result.Decode(val)
	if err != nil {
		return "", err
	}
	element, err := val.Lookup("value")
	if err != nil {
		return "", err
	}
	return element.Value().StringValue(), nil
}

func (m *MongoDB) FindAll() ([]Document, error) {
	var result []Document
	cursor, err := m.activeCollection.Find(context.Background(), bson.NewDocument())
	if err != nil {
		return result, err
	}
	for cursor.Next(context.Background()) {
		val, err := m.decodeValue(cursor)
		if err != nil {
			return result, err
		}
		result = append(result, val)
	}
	return result, nil
}

func (m *MongoDB) decodeValue(cur mongo.Cursor) (Document, error) {
	val := bson.NewDocument()

	err := cur.Decode(val)
	if err != nil {
		return Document{}, err
	}
	element, err := val.Lookup("value")
	if err != nil {
		return Document{}, err
	}
	value := element.Value().StringValue()

	element, err = val.Lookup("key")
	if err != nil {
		return Document{}, err
	}
	key := element.Value().StringValue()
	return Document{Key: key, Value: value}, nil
}

func (m *MongoDB) InsertOneSimple(key, value string) error {
	_, err := m.activeCollection.InsertOne(context.Background(), bson.NewDocument(
		bson.EC.String("key", key),
		bson.EC.String("value", value),
	))
	return err
}

type Document struct {
	Key, Value string
}
