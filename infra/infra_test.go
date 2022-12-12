package infra

import (
	"context"
	"log"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	// MongoDB の URI（テスト用）
	mongoURI = "mongodb://root:example@localhost:27017"
)

func TestMain(m *testing.M) {
	// MongoDB との疎通確認
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Panicf("MongoDB への接続に失敗しました: %v", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	_ = client.Disconnect(ctx)
	if err != nil {
		log.Panicf("MongoDB への接続に失敗しました: %v", err)
	}

	m.Run()
}
