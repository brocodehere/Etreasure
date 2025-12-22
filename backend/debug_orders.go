package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to database
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/etreasure?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test query to see what fields exist
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_name = 'orders' 
		ORDER BY ordinal_position
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Orders table schema:")
	fmt.Println("====================")

	for rows.Next() {
		var columnName, dataType, isNullable string
		err := rows.Scan(&columnName, &dataType, &isNullable)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%-20s %-15s %-8s\n", columnName, dataType, isNullable)
	}

	// Test if we can select the price fields
	fmt.Println("\nTesting price field selection:")
	fmt.Println("===========================")

	testQuery := `
		SELECT 
			id, order_number, 
			COALESCE(total_price, 0) as total_price,
			COALESCE(subtotal, 0) as subtotal,
			COALESCE(tax_amount, 0) as tax_amount,
			COALESCE(shipping_amount, 0) as shipping_amount,
			COALESCE(discount_amount, 0) as discount_amount
		FROM orders 
		LIMIT 1
	`

	rows, err = db.Query(testQuery)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	defer rows.Close()

	fmt.Println("Query executed successfully!")

	for rows.Next() {
		var id, orderNumber string
		var totalPrice, subtotal, taxAmount, shippingAmount, discountAmount float64
		err := rows.Scan(&id, &orderNumber, &totalPrice, &subtotal, &taxAmount, &shippingAmount, &discountAmount)
		if err != nil {
			fmt.Printf("Scan ERROR: %v\n", err)
			return
		}
		fmt.Printf("ID: %s, Total: %.2f, Subtotal: %.2f\n", id, totalPrice, subtotal)
		break
	}
}
