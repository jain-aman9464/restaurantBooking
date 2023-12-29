package services

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"restaurentBooking/src/database"
	"restaurentBooking/src/model"
)

func CreateRestaurant(c *gin.Context) {
	var restaurant model.Restaurant
	if err := c.BindJSON(&restaurant); err != nil {
		fmt.Println("err: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant data"})
		return
	}

	// Save the restaurant to the database
	var restID int
	var err error
	if restID, err = database.SaveRestaurant(&restaurant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create restaurant"})
		return
	}

	restaurant.ID = restID
	c.JSON(http.StatusCreated, restaurant)
}

func GetRestaurants(c *gin.Context) {
	restaurants, err := database.GetRestaurantsFromDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"restaurants": restaurants})
}
