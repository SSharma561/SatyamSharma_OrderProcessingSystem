# SatyamSharma_OrderProcessingSystem

# Order Service
This is a simple REST API built with Go, Gin, and Gorm for creating orders.

## Prerequisites
Golang 1.20+
PostgreSQl 14+
go, psql CLI tools

## Installation
git clone https://github.com/ssharma561/SatyamSharma_OrderProcessingSystem.git
cd SatyamSharma_OrderProcessingSystem

## Install Dependencies
go mod tidy

## How to Run
go run main.go

## Server Runing On
http://localhost:8080

## API Endpoints
GET /api/customers - Get all customers
GET /api/customers/{id} - Get a specific customer by ID Including Order Details
POST /api/orders - Create a new order
    --Request Body [JSON]
    {
        "customer_id": 2,
        "product_ids": [1, 2, 3]
    }
GET /api/orders/{id} - Get details of a specific order including total price


