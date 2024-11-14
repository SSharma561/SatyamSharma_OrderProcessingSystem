package main

import (
	"order_processing_system/controllers"
	"order_processing_system/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	database.InitDB()

	r := gin.Default()

	// Define API routes
	r.GET("/api/customers", controllers.GetCustomers)
	r.GET("/api/customers/:id", controllers.GetCustomerByID)
	r.POST("/api/orders", controllers.CreateOrder)
	r.GET("/api/orders/:id", controllers.GetOrderByID)

	// Start the server
	r.Run("localhost:8080")
}
