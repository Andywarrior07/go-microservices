package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

var channelAmqp *amqp.Channel

type Request struct {
	URL string `json:"url"`
}

func init() {
	amqpConnection, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		log.Fatal(err)
	}

	channelAmqp, _ = amqpConnection.Channel()
}

func ParserHandler(c *gin.Context) {
	var request Request

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	data, _ := json.Marshal(request)

	err := channelAmqp.Publish(
		"",
		"rss_urls",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		},
	)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error while publishing to RabbitMQ",
		})

		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"message": "Success",
	})
}

func main() {
	router := gin.Default()

	router.POST("/parse", ParserHandler)

	router.Run(":5000")
}
