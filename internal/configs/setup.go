package configs

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(EnvMongoURI()))
	if err != nil {
		log.Error().Err(err).Msg("Error creating mongo client")
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Error connecting to mongo")
	}

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Error pinging mongo")
	}
	log.Info().Msg("Connected to MongoDB!")
	return client
}

// Client instance
var DB *mongo.Client = ConnectDB()

// getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("project").Collection(collectionName)
	return collection
}
