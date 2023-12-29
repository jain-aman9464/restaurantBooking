package mq

import (
	"log"

	"github.com/streadway/amqp"
)

var rabbitMQURL = "amqp://guest:guest@localhost:5672/"

func CreateRabbitMQConnection() (*amqp.Connection, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	return conn, err
}
