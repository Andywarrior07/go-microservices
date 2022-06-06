package main

import (
	"context"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var ctx context.Context

type Feed struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Thumbnail struct {
		Url string `xml:"url,attr"`
	} `xml:"thumbnail"`
	Title string `xml:"title"`
}

type Request struct {
	URL string `json:"url"`
}

func getMongoUri() string {
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv("MONGO_URI")
}

func GetFeedEntries(url string) ([]Entry, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0;Win64;x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	byteValue, _ := ioutil.ReadAll(resp.Body)

	var feed Feed

	xml.Unmarshal(byteValue, &feed)

	return feed.Entries, nil
}

func ParserHandler(c *gin.Context) {
	var request Request

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	entries, err := GetFeedEntries(request.URL)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while parsing the RSS feed",
		})

		return
	}

	c.JSON(http.StatusOK, entries)
}

func init() {
	ctx = context.Background()
	client, _ = mongo.Connect(ctx, options.Client().ApplyURI(getMongoUri()))
}

func main() {
	router := gin.Default()

	router.POST("/parse", ParserHandler)

	router.Run(":5000")
}
