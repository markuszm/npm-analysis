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

func (m *MongoDB) InsertOneSimple(key, value string) error {
	_, err := m.activeCollection.InsertOne(context.Background(), bson.NewDocument(
		bson.EC.String("key", key),
		bson.EC.String("value", value),
	))
	return err
}
