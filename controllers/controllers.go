package controllers

import (
	"net/http"
	"order_processing_system/database"
	"order_processing_system/models"

	"github.com/gin-gonic/gin"
)

// Get all customers
func GetCustomers(c *gin.Context) {
	var customers []models.Customer
	query := `
		SELECT * FROM customers;
	`
	result := database.DB.Raw(query).Scan(&customers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

// Get a specific customer by ID including their orders
func GetCustomerByID(c *gin.Context) {
	customerID := c.Param("id")
	var customer models.Customer
	var orders []models.Order

	// Fetch Customer
	customerResult := database.DB.Raw("SELECT * FROM customers WHERE id = ?", customerID).Scan(&customer)
	if customerResult.Error != nil || customer.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Fetch Orders
	orderResult := database.DB.Raw("SELECT * FROM orders WHERE customer_id = ?", customerID).Scan(&orders)
	if orderResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": orderResult.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer": customer,
		"orders":   orders,
	})
}

// Create a new order for a customer
func CreateOrder(c *gin.Context) {
	var request struct {
		CustomerID uint   `json:"customer_id"`
		ProductIDs []uint `json:"product_ids"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx := database.DB.Begin()

	// Check if customer exists
	var customer models.Customer
	customerResult := tx.Raw("SELECT * FROM customers WHERE id = ?", request.CustomerID).Scan(&customer)
	if customerResult.Error != nil || customer.ID == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Check if the customer has any unfulfilled orders
	var existingOrder models.Order
	unfulfilledOrderCheck := tx.Raw("SELECT * FROM orders WHERE customer_id = ? AND status = 'unfulfilled' LIMIT 1", request.CustomerID).Scan(&existingOrder)
	if unfulfilledOrderCheck.Error == nil && existingOrder.ID != 0 {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot place a new order. Previous order is unfulfilled"})
		return
	}

	// Calculate total price
	query := `SELECT SUM(price) as total_price FROM products WHERE id IN (?);`
	var totalPrice float64
	err := tx.Raw(query, request.ProductIDs).Scan(&totalPrice).Error
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
		return
	}

	// Create Order
	order := models.Order{
		CustomerID: request.CustomerID,
		TotalPrice: totalPrice,
		Status:     "unfulfilled",
	}
	createOrder := tx.Raw("INSERT INTO orders (customer_id, total_price, status) VALUES (?, ?, ?) RETURNING id", order.CustomerID, order.TotalPrice, order.Status).Scan(&order.ID)
	if createOrder.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": createOrder.Error.Error()})
		return
	}

	// Insert Order Products
	for _, productID := range request.ProductIDs {
		err := tx.Exec("INSERT INTO order_products (order_id, product_id) VALUES (?, ?)", order.ID, productID).Error
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusCreated, order)
}

// Get Order by ID
func GetOrderByID(c *gin.Context) {
	orderID := c.Param("id")
	var order models.Order

	// Fetch order
	orderResult := database.DB.Raw("SELECT * FROM orders WHERE id = ?", orderID).Scan(&order)
	if orderResult.Error != nil || order.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	var products []models.Product
	err := database.DB.Raw(`
		SELECT p.id, p.name, p.price
		FROM products p
		JOIN order_products op ON op.product_id = p.id
		WHERE op.order_id = ?`, orderID).Scan(&products).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order":    order,
		"products": products,
	})
}
