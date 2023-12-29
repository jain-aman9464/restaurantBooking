package main

import (
	"encoding/json"
	"fmt"
	"log"
	"restaurentBooking/src/mq"
)

func main() {
	conn, err := mq.CreateRabbitMQConnection()
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	// Declare an exchange (matching the one used for publishing)
	err = ch.ExchangeDeclare(
		"reservation_exchange", // exchange name
		"fanout",               // exchange type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatal("Failed to declare an exchange:", err)
	}

	// Declare a queue (you may need to adapt this based on your RabbitMQ setup)
	q, err := ch.QueueDeclare(
		"email_queue", // queue name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err)
	}

	// Bind the queue to the exchange
	err = ch.QueueBind(
		q.Name,                 // queue name
		"",                     // routing key (empty for fanout)
		"reservation_exchange", // exchange
		false,                  // no-wait
		nil,                    // arguments
	)
	if err != nil {
		log.Fatal("Failed to bind the queue to the exchange:", err)
	}

	// Consume messages from the queue
	msgs, err := ch.Consume(
		q.Name, // queue name
		"",     // consumer
		true,   // auto-acknowledge
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	// Process incoming messages
	for msg := range msgs {
		// Extract user and reservation details from the JSON message body
		var messageBody map[string]interface{}
		if err := json.Unmarshal(msg.Body, &messageBody); err != nil {
			log.Println("Error decoding message body:", err)
			continue
		}

		// Extract user and reservation details
		user := messageBody["user"].(map[string]interface{})
		reservation := messageBody["reservation"].(map[string]interface{})
		fmt.Println("user: ", user)
		fmt.Println("reservation: ", reservation)
		// Now you have user and reservation details for email notification
		userEmail := user["user_email"].(string)
		userName := user["user_name"].(string)
		reservationID := int(reservation["reservation_id"].(float64))
		reservationFromTime := reservation["reservation_from"].(string)
		reservationToTime := reservation["reservation_to"].(string)

		// TODO: Replace the following line with your actual email sending logic
		sendEmailNotification(userEmail, userName, reservationID, reservationFromTime, reservationToTime)
	}
}

// Placeholder function for sending email notification (replace with actual logic)
func sendEmailNotification(userEmail, userName string, reservationID int, reservationTime, reservationToTime string) {
	log.Printf("Sending email notification to %s for reservation %d at %s\n", userEmail, reservationID, reservationTime)
}
