// Blogo API
//
// This is the API implementation for Blogo
//
//  Schemes: http
//  Host: localhost:8080
//  BasePath: /
//  Version: 1.0.0
//  Contact: Xiaoxi Sun <xiaoxisun2000@gmail.com>
//
//  Consumes:
//  - application/json
//
//  Produces:
//  - application/json
// swagger:meta
package main

import (
	"blogo/handlers"
	"context"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var postsHandlers *handlers.PostsHandler

func init() {
	ctx := context.Background()
	// Connect to Mongo
	mongo_uri := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongo_uri))
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	collectionPosts := client.Database(os.Getenv("MONGO_DATABASE")).Collection("posts")
	log.Println(os.Getenv("MONGO_DATABASE"))
	// collectionPosts := client.Database("blogo").Collection("posts")
	// Connect to redis
	log.Println(collectionPosts)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping()
	log.Println(status)

	postsHandlers = handlers.NewPostsHandlers(ctx, collectionPosts, redisClient)
}

func main() {
	router := gin.Default()

	// no auth
	router.GET("/posts", postsHandlers.ListPostsHandler)
	router.GET("/posts/:id", postsHandlers.ViewPostHandler)
	router.GET("/posts/search/:title", postsHandlers.SearchPostHandler)

	// need auth
	router.DELETE("/posts/:id", postsHandlers.DeletePostHandler)
	router.POST("/posts", postsHandlers.NewPostHandler)
	router.POST("/posts/thumbup/:id", postsHandlers.ThumbupPostHandler)
	router.Run()
}
