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
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	sessionRedisStore "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var postsHandlers *handlers.PostsHandler
var commentsHandlers *handlers.CommentsHandler
var authhandler *handlers.AuthHandler

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
	collectionComments := client.Database(os.Getenv("MONGO_DATABASE")).Collection("comments")
	collectionUsers := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")

	// Connect to redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	status := redisClient.Ping()
	log.Println(status)
	//create handlers
	postsHandlers = handlers.NewPostsHandlers(ctx, collectionPosts, redisClient)
	commentsHandlers = handlers.NewCommentsHandlers(ctx, collectionComments, redisClient)
	authhandler = handlers.NewAuthHandler(ctx, collectionUsers)
}

func main() {
	router := gin.Default()

	store, _ := sessionRedisStore.NewStore(10, "tcp", os.Getenv("SESSION_REDIS_URI"), "", []byte("secret"))

	router.Use(sessions.Sessions("post_api", store))
	corsConfig := cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000/write-post", "http://localhost:3000"},
		AllowMethods:     []string{"POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Length", "Content-Type", "Accept", "Authorization", "Access-Control-Request-Credentials", "Access-Control-Request-Origin", "Access-Control-Request-Methods"},
		ExposeHeaders:    []string{"Cookie"},
		AllowCredentials: true,
		MaxAge:           60 * 60 * time.Hour,
	})

	// corsConfig := cors.Default()
	router.Use(corsConfig)

	// sign in
	router.POST("/signin", authhandler.SignInHandler)
	router.POST("/singout", authhandler.SignOutHandler)
	router.POST("/signup", authhandler.SignUpHandler)

	// view posts
	router.GET("/posts", postsHandlers.ListPostsHandler)
	router.GET("/posts/:id", postsHandlers.ViewPostHandler)
	router.GET("/posts/search/:title", postsHandlers.SearchPostHandler)
	router.GET("/random-post", postsHandlers.GetOneRandomPost)

	// view comments
	router.GET("/comments/:postid", commentsHandlers.ListCommentsToPostHandler)
	authorized := router.Group("/")

	// newCorsConfig := cors.New(cors.Config{
	// 	AllowOrigins:     []string{"http://localhost:3000"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization"},
	// 	AllowCredentials: true,
	// 	MaxAge:           12 * time.Hour,
	// })
	// authorized.Use(cors.Default())

	authorized.Use(corsConfig)

	authorized.Use(authhandler.AuthMiddileware())
	{
		authorized.DELETE("/posts/:id", postsHandlers.DeletePostHandler)
		authorized.POST("/posts", postsHandlers.NewPostHandler)
		authorized.POST("/posts/thumbup/:id", postsHandlers.ThumbupPostHandler)
		authorized.POST("/comments/:postid", commentsHandlers.CreateCommentToPostHandler)
		authorized.POST("/comments/thumbup/:commentid", commentsHandlers.CommentThumbupHandler)
	}

	router.Run()
}
