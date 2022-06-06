package main

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Request struct {
	URL string `json:"url"`
}

type Feed struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Thumbnail struct {
		URL string `xml:"url,attr"`
	} `xml:"thumbnail"`
	Title string `xml:"title"`
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

func main() {
	ctx := context.Background()

	mongoClient, _ := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017/recipesdb"))

	defer mongoClient.Disconnect(ctx)

	amqpConnection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		log.Fatal(err)
	}

	defer amqpConnection.Close()

	channelAmqp, _ := amqpConnection.Channel()

	defer channelAmqp.Close()

	forever := make(chan bool)

	msgs, err := channelAmqp.Consume(
		"rss_urls",
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			var request Request

			json.Unmarshal(d.Body, &request)

			log.Println("RSS URL: ", request.URL)

			entries, _ := GetFeedEntries(request.URL)

			collection := mongoClient.Database("recipesdb").Collection("recipes")

			for _, entry := range entries[2:] {
				collection.InsertOne(ctx, bson.M{
					"title":     entry.Title,
					"thumbnail": entry.Thumbnail.URL,
					"url":       entry.Link.Href,
				})
			}
		}
	}()

	log.Printf("[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
