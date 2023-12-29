package services

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"restaurentBooking/src/database"
	"restaurentBooking/src/model"
	"strconv"
)

func CreateUser(c *gin.Context) {
	var user model.User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data"})
		return
	}

	var userID int
	var err error
	// Save the user to the database
	if userID, err = database.SaveUser(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	user.ID = userID
	c.JSON(http.StatusCreated, user)
}

func GetUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := database.GetUserFromDB(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}
