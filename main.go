package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"restaurentBooking/src/database"
	"restaurentBooking/src/services"
)

func main() {
	router := gin.Default()

	err := database.InitDB()
	if err != nil {
		log.Fatal("Db cannot be initialized: ", err)
	}

	router.POST("/createReservation", services.CreateReservation)
	router.POST("/createRestaurant", services.CreateRestaurant)
	router.POST("/createUser", services.CreateUser)

	router.GET("/getRestaurants", services.GetRestaurants)
	router.GET("/getUser", services.GetUser)
	router.GET("/getAvailableTables", services.GetAvailableTablesHandler)

	err = router.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
