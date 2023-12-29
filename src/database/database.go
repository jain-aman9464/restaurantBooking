package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/streadway/amqp"
	"log"
	"restaurentBooking/src/model"
	"restaurentBooking/src/mq"
	"time"
)

var db *sql.DB

// InitDB initializes the database connection
func InitDB() error {
	var err error
	db, err = sql.Open("mysql", "root:asu@123asu@123@tcp(localhost:3306)/db_reservations")
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Ensure the necessary tables exist and perform any other database initialization

	return nil
}

func GetUserFromDB(userID int) (model.User, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return model.User{}, err
	}
	defer tx.Rollback()

	var name, email string
	var age int
	err = tx.QueryRow("SELECT name, age, email FROM users WHERE id = ?", userID).Scan(&name, &age, &email)
	if err != nil {
		log.Println("Error fetching user:", err)
		return model.User{}, err
	}

	user := model.User{
		ID:    userID,
		Name:  name,
		Age:   age,
		Email: email,
	}

	return user, nil
}

// SaveUser saves a user to the database
func SaveUser(user *model.User) (int, error) {
	result, err := db.Exec("INSERT INTO users (name, age, email) VALUES (?,?,?)", user.Name, user.Age, user.Email)
	if err != nil {
		log.Println("Error saving user:", err)
		return 0, err
	}

	lastInsertID, _ := result.LastInsertId()

	return int(lastInsertID), nil
}

func GetRestaurantsFromDB() ([]model.Restaurant, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback()

	rows, err := tx.Query("SELECT id, name, cuisine FROM restaurants")
	if err != nil {
		log.Println("Error fetching restaurants:", err)
		return nil, err
	}
	defer rows.Close()

	var restaurants []model.Restaurant
	for rows.Next() {
		var restaurantID int
		var name, cuisine, location string

		err := rows.Scan(&restaurantID, &name, &cuisine, &location)
		if err != nil {
			log.Println("Error scanning restaurant rows:", err)
			return nil, err
		}
		restaurant := model.Restaurant{
			ID:       restaurantID,
			Name:     name,
			Cuisine:  cuisine,
			Location: location,
		}

		restaurants = append(restaurants, restaurant)
	}

	return restaurants, nil
}

// SaveRestaurant saves a restaurant to the database
func SaveRestaurant(restaurant *model.Restaurant) (int, error) {
	result, err := db.Exec("INSERT INTO restaurants (name, cuisine, location) VALUES (?, ?, ?)", restaurant.Name, restaurant.Cuisine, restaurant.Location)
	if err != nil {
		log.Println("Error saving restaurant:", err)
		return 0, err
	}

	lastInsertID, _ := result.LastInsertId()
	return int(lastInsertID), nil
}

// SaveReservation saves a reservation to the database
func SaveReservation(tableID, userID, restaurantID int, reservationFrom, reservationTo time.Time) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	// Check table availability with FOR UPDATE
	var status string
	err = tx.QueryRow("SELECT status FROM reservations WHERE table_id = ? AND reservation_to > ? AND reservation_from < ? FOR UPDATE",
		tableID, reservationFrom, reservationTo).Scan(&status)
	if err == sql.ErrNoRows {
		// Handle the case where no rows are found
		// For example, set a default status or return an error
		log.Println("No matching reservation found")
		status = "available"
		// Handle the error or set a default status
		// ...
	} else if err != nil {
		log.Println("Error checking table availability:", err)
		return err
	}

	if status != "available" {
		// Table is not available for the specified time range
		return errors.New("table not available for the specified time range")
	}

	// Insert the reservation details into the reservations table
	res, err := tx.Exec("INSERT INTO reservations (user_id, restaurant_id, table_id, reservation_from, reservation_to, status) VALUES (?, ?, ?, ?, ?, ?)",
		userID, restaurantID, tableID, reservationFrom, reservationTo, "booked")
	if err != nil {
		log.Println("Error making reservation:", err)
		return err
	}

	// Update the status of the table in the restaurant_tables table
	_, err = tx.Exec("UPDATE restaurant_tables SET status = ? WHERE id = ?", "booked", tableID)
	if err != nil {
		log.Println("Error updating table status in restaurant_tables:", err)
		return err
	}

	conn, err := mq.CreateRabbitMQConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

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
		return err
	}

	user, _ := GetUserFromDB(userID)
	fmt.Println("user: ", user)
	reservationID, _ := res.LastInsertId()
	userDetails := fmt.Sprintf(`{"user_id": %d, "user_name": "%s", "user_email": "%s"}`, user.ID, user.Name, user.Email)
	fmt.Println("userdetails: ", userDetails)
	reservationDetails := fmt.Sprintf(`{"reservation_id": %d, "reservation_from": "%s", "reservation_to": "%s"}`, int(reservationID), reservationFrom, reservationTo)

	// Combine user and reservation details into a single message body
	messageBody := fmt.Sprintf(`{"user": %s, "reservation": %s}`, userDetails, reservationDetails)

	// Publish a message to the exchange
	err = ch.Publish(
		"reservation_exchange", // exchange
		"",                     // routing key (empty for fanout)
		false,                  // mandatory
		false,                  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(messageBody),
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// getAllAvailableTables returns a list of all available tables for a given restaurant and time range
func GetAllAvailableTables(restaurantID int, fromTime, toTime time.Time) ([]gin.H, error) {
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return nil, err
	}
	defer tx.Rollback()

	// Query all tables for the specified restaurant
	rows, err := tx.Query("SELECT id, table_number, capacity FROM restaurant_tables WHERE restaurant_id = ?", restaurantID)
	if err != nil {
		log.Println("Error fetching tables:", err)
		return nil, err
	}
	defer rows.Close()

	// Create a map for each available table
	var availableTables []gin.H
	for rows.Next() {
		var tableID, tableNumber, capacity int

		err := rows.Scan(&tableID, &tableNumber, &capacity)
		if err != nil {
			log.Println("Error scanning table rows:", err)
			return nil, err
		}

		// Check if the table is available within the specified time range
		if isTableAvailable(tableID, fromTime, toTime) {
			availableTable := gin.H{
				"id":       tableID,
				"number":   tableNumber,
				"capacity": capacity,
				// Add other table details if needed
			}
			availableTables = append(availableTables, availableTable)
		}
	}

	return availableTables, nil
}

// isTableAvailable checks if the specified table is available within the given time range
func isTableAvailable(tableID int, fromTime, toTime time.Time) bool {
	// Implement your logic to check table availability within the specified time range
	// For example, query the reservations table for conflicting reservations
	// Return true if the table is available, false otherwise
	// You may need to modify this function based on your specific requirements
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		log.Println("Error starting transaction:", err)
		return false
	}
	defer tx.Rollback()

	// Query for reservations that overlap with the specified time range
	rows, err := tx.Query("SELECT id FROM reservations WHERE table_id = ? AND reservation_to > ? AND reservation_from < ?", tableID, fromTime, toTime)
	if err == sql.ErrNoRows {
		// Handle the case where no rows are found
		// For example, set a default status or return an error
		log.Println("No matching reservation found")
		// Handle the error or set a default status
		// ...
	} else if err != nil {
		log.Println("Error checking table availability:", err)
		return false
	}
	defer rows.Close()

	// If there are overlapping reservations, the table is not available
	return !rows.Next()
}
