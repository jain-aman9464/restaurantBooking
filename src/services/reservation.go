package services

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"restaurentBooking/src/database"
	"strconv"
	"time"
)

func GetAvailableTablesHandler(c *gin.Context) {
	// Parse parameters from the request
	restaurantID := c.Query("restaurant_id")
	fromTimeStr := c.Query("from_time")
	toTimeStr := c.Query("to_time")
	restaurantIDInt, _ := strconv.Atoi(restaurantID)

	// Convert time strings to time.Time objects
	fromTime, err := time.Parse("2006-01-02T15:04:05", fromTimeStr)
	if err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from_time format"})
		return
	}

	toTime, err := time.Parse("2006-01-02T15:04:05", toTimeStr)
	if err != nil {
		fmt.Println("err:===", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to_time format"})
		return
	}

	// Call the function to get available tables
	availableTables, err := database.GetAllAvailableTables(restaurantIDInt, fromTime, toTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Respond with the list of available tables
	c.JSON(http.StatusOK, gin.H{"tables": availableTables})
}

func CreateReservation(c *gin.Context) {
	tableID, err := strconv.Atoi(c.PostForm("table_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table ID"})
		return
	}

	userID, errUsr := strconv.Atoi(c.PostForm("user_id"))
	if errUsr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	restaurantID, errRest := strconv.Atoi(c.PostForm("restaurant_id"))
	if errRest != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID"})
		return
	}

	fromTimeStr := c.PostForm("reservation_from")
	toTimeStr := c.PostForm("reservation_to")

	// Parse the date_time string to a time.Time object
	fromTime, errDate := time.Parse("2006-01-02T15:04:05", fromTimeStr)
	if errDate != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation_from format"})
		return
	}

	toTime, errDate := time.Parse("2006-01-02T15:04:05", toTimeStr)
	if errDate != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reservation_to format"})
		return
	}
	// Save the reservation to the database
	if err = database.SaveReservation(tableID, userID, restaurantID, fromTime, toTime); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "reservation created"})
	return
}
