package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alpkeskin/gotoon"
)

func main() {
	// Example 1: Simple object
	fmt.Println("=== Example 1: Simple Object ===")
	simpleData := map[string]interface{}{
		"id":     123,
		"name":   "Ada",
		"active": true,
	}
	encoded, err := gotoon.Encode(simpleData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 2: Tabular array (uniform objects)
	fmt.Println("=== Example 2: Tabular Array ===")
	usersData := map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "Alice", "role": "admin"},
			{"id": 2, "name": "Bob", "role": "user"},
			{"id": 3, "name": "Charlie", "role": "user"},
		},
	}
	encoded, err = gotoon.Encode(usersData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 3: Nested structures
	fmt.Println("=== Example 3: Nested Structures ===")
	nestedData := map[string]interface{}{
		"user": map[string]interface{}{
			"id":   456,
			"name": "Eve",
			"preferences": map[string]interface{}{
				"theme":    "dark",
				"language": "en",
			},
		},
	}
	encoded, err = gotoon.Encode(nestedData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 4: Primitive arrays
	fmt.Println("=== Example 4: Primitive Arrays ===")
	arrayData := map[string]interface{}{
		"tags":   []string{"reading", "gaming", "coding"},
		"scores": []int{85, 92, 78, 95},
	}
	encoded, err = gotoon.Encode(arrayData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 5: Using structs
	fmt.Println("=== Example 5: Using Structs ===")
	type Product struct {
		SKU   string  `json:"sku"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
		Stock int     `json:"stock"`
	}

	productsData := map[string]interface{}{
		"products": []Product{
			{SKU: "A001", Name: "Widget", Price: 19.99, Stock: 100},
			{SKU: "A002", Name: "Gadget", Price: 29.99, Stock: 50},
			{SKU: "A003", Name: "Doohickey", Price: 9.99, Stock: 200},
		},
	}
	encoded, err = gotoon.Encode(productsData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 6: With custom options
	fmt.Println("=== Example 6: Tab Delimiter ===")
	encoded, err = gotoon.Encode(usersData, gotoon.WithDelimiter("\t"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 7: With length marker
	fmt.Println("=== Example 7: Length Marker ===")
	encoded, err = gotoon.Encode(arrayData, gotoon.WithLengthMarker())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 8: Mixed array
	fmt.Println("=== Example 8: Mixed Array ===")
	mixedData := map[string]interface{}{
		"items": []interface{}{
			1,
			"text",
			true,
			map[string]interface{}{"key": "value"},
		},
	}
	encoded, err = gotoon.Encode(mixedData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 9: Time values
	fmt.Println("=== Example 9: Time Values ===")
	timeData := map[string]interface{}{
		"event": "User Login",
		"timestamp": time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	encoded, err = gotoon.Encode(timeData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
	fmt.Println()

	// Example 10: E-commerce order
	fmt.Println("=== Example 10: E-commerce Order ===")
	orderData := map[string]interface{}{
		"order": map[string]interface{}{
			"id":     "ORD-12345",
			"date":   "2025-01-15",
			"status": "shipped",
			"customer": map[string]interface{}{
				"name":  "John Doe",
				"email": "john@example.com",
			},
			"items": []map[string]interface{}{
				{"sku": "WIDGET-1", "quantity": 2, "price": 19.99},
				{"sku": "GADGET-2", "quantity": 1, "price": 49.99},
			},
			"total": 89.97,
		},
	}
	encoded, err = gotoon.Encode(orderData)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(encoded)
}
