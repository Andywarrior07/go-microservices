package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var collection *mongo.Collection
var ctx context.Context

type Recipe struct {
	Title     string `json:"title" bson:"title"`
	Thumbnail string `json:"thumbnail" bson:"thumbnail"`
	URL       string `json:"url" bson:"url"`
}

func IndexHandler(c *gin.Context) {
	cur, err := collection.Find(ctx, bson.M{})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer cur.Close(ctx)

	recipes := make([]Recipe, 0)

	for cur.Next(ctx) {
		var recipe Recipe
		cur.Decode(&recipe)

		recipes = append(recipes, recipe)
	}

	c.HTML(http.StatusOK, "index.tmlp", gin.H{
		"recipes": recipes,
	})

}

func init() {
	ctx = context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/recipesdb"))

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to mongoDB")
	collection = client.Database("recipesdb").Collection("recipes")
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")
	router.GET("/dashboard", IndexHandler)
	router.Run(":3000")
}
