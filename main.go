package main

import (
	"context"
	"net/http"

	"nekonoshiri/go-echo-sample/infra"
	"nekonoshiri/go-echo-sample/usecase"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	// MongoDB の URI
	mongoURI = "mongodb://root:example@mongo:27017"
)

func main() {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("MongoDB への接続に失敗しました: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("MongoDB からの切断時にエラーが発生しました: %v", err)
		}
	}()

	userRepository := infra.NewMongoUserRepository(client)

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Logger.SetLevel(log.INFO)

	e.GET("/users/:userID", func(c echo.Context) error {
		return usecase.GetUser(c, userRepository)
	})

	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatalf("サーバーにエラーが発生しました: %v", err)
	}
}
