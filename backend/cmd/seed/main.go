package main

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}
	defer pool.Close()

	if err := seedRolesAndAdmin(ctx, pool); err != nil {
		log.Fatalf("seed roles/admin failed: %v", err)
	}

	if err := seedCategories(ctx, pool); err != nil {
		log.Fatalf("seed categories failed: %v", err)
	}

	if err := seedCartAndWishlist(ctx, pool); err != nil {
		log.Fatalf("seed cart/wishlist failed: %v", err)
	}

	log.Println("seed completed - admin users, categories, cart and wishlist created")
}

func seedRolesAndAdmin(ctx context.Context, pool *pgxpool.Pool) error {
	roles := []string{"SuperAdmin", "Admin", "Editor", "Support"}
	for _, r := range roles {
		_, err := pool.Exec(ctx, `INSERT INTO roles (name) VALUES ($1) ON CONFLICT (name) DO NOTHING`, r)
		if err != nil {
			return err
		}
	}

	// Create users for each role
	users := []struct {
		email    string
		fullName string
		role     string
		password string
	}{
		{
			email:    "superadmin@etreasure.com",
			fullName: "Super Admin User",
			role:     "SuperAdmin",
			password: "SuperAdmin123!",
		},
		{
			email:    "admin@etreasure.com",
			fullName: "Admin User",
			role:     "Admin",
			password: "Admin123!",
		},
		{
			email:    "editor@etreasure.com",
			fullName: "Editor User",
			role:     "Editor",
			password: "Editor123!",
		},
		{
			email:    "support@etreasure.com",
			fullName: "Support User",
			role:     "Support",
			password: "Support123!",
		},
	}

	for _, user := range users {
		hash, err := bcrypt.GenerateFromPassword([]byte(user.password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		var userID int
		err = pool.QueryRow(ctx,
			`INSERT INTO users (email, password_hash, full_name, is_active) 
			 VALUES ($1, $2, $3, TRUE)
			 ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash, full_name = EXCLUDED.full_name
			 RETURNING id`,
			user.email, string(hash), user.fullName).Scan(&userID)
		if err != nil {
			return err
		}

		_, err = pool.Exec(ctx, `INSERT INTO user_roles (user_id, role_id)
			SELECT $1, id FROM roles WHERE name = $2
			ON CONFLICT (user_id, role_id) DO NOTHING`, userID, user.role)
		if err != nil {
			return err
		}

		log.Printf("Created user: %s with role: %s (password: %s)", user.email, user.role, user.password)
	}

	// Also create the original admin@example.com for backward compatibility
	password := os.Getenv("ADMIN_INITIAL_PASSWORD")
	if password == "" {
		password = "ChangeMe123!"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	var userID int
	err = pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, full_name, is_active) 
		 VALUES ($1, $2, $3, TRUE)
		 ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
		 RETURNING id`,
		"admin@example.com", string(hash), "Initial Admin").Scan(&userID)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'SuperAdmin'
		ON CONFLICT (user_id, role_id) DO NOTHING`, userID)
	if err != nil {
		return err
	}

	log.Printf("Created legacy admin: admin@example.com with role: SuperAdmin (password: %s)", password)

	return nil
}

func seedCategories(ctx context.Context, pool *pgxpool.Pool) error {
	categories := []struct {
		slug        string
		name        string
		description *string
		sortOrder   int
	}{
		{"hand-bags", "Hand Bags", stringPtr("Premium leather and crafted hand bags"), 1},
		{"laptop-bags", "Laptop Bags", stringPtr("Professional and stylish laptop bags"), 2},
		{"tote-bags", "Tote Bags", stringPtr("Spacious and eco-friendly tote bags"), 3},
		{"crochet-throws", "Crochet Throws", stringPtr("Handmade crochet throws for home decor"), 4},
		{"embroidered-clutches", "Embroidered Clutches", stringPtr("Elegant embroidered clutches for special occasions"), 5},
		{"baby-throw", "Baby Throw", stringPtr("Soft and safe throws for babies"), 6},
		{"everyday-bags", "Everyday Bags", stringPtr("Durable bags for daily use"), 7},
	}

	for _, cat := range categories {
		_, err := pool.Exec(ctx, `
			INSERT INTO categories (slug, name, description, sort_order) 
			VALUES ($1, $2, $3, $4) 
			ON CONFLICT (slug) DO NOTHING`,
			cat.slug, cat.name, cat.description, cat.sortOrder)
		if err != nil {
			return err
		}
	}

	log.Println("Created sample categories")
	return nil
}

func seedCartAndWishlist(ctx context.Context, pool *pgxpool.Pool) error {
	// Debug: Check customers table structure first
	log.Printf("Checking customers table structure...")
	var customerCount int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM customers").Scan(&customerCount)
	if err != nil {
		log.Printf("Error checking customers count: %v", err)
		return err
	}
	log.Printf("Found %d customers in customers table", customerCount)

	// First, create a sample customer if it doesn't exist
	var customerID uuid.UUID
	err = pool.QueryRow(ctx, `
		INSERT INTO customers (email, first_name, last_name) 
		VALUES ('demo@etreasure.com', 'Demo', 'User')
		ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name, last_name = EXCLUDED.last_name
		RETURNING id
	`).Scan(&customerID)
	if err != nil {
		log.Printf("Error creating customer: %v", err)
		return err
	}
	log.Printf("Created customer with ID: %s", customerID.String())

	// Debug: Check products table structure first
	log.Printf("Checking products table structure...")
	var productCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&productCount)
	if err != nil {
		log.Printf("Error checking products count: %v", err)
		return err
	}
	log.Printf("Found %d products in products table", productCount)

	// Try to get product IDs - check if they use id or uuid_id
	rows, err := pool.Query(ctx, `
		SELECT uuid_id FROM products 
		WHERE published = TRUE 
		ORDER BY created_at DESC 
		LIMIT 5
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var productIDs []uuid.UUID
	for rows.Next() {
		var uuidID uuid.UUID
		if err := rows.Scan(&uuidID); err != nil {
			log.Printf("Error scanning product row: %v", err)
			return err
		}
		productIDs = append(productIDs, uuidID)
		log.Printf("Found product: uuid_id=%s", uuidID.String())
	}

	if len(productIDs) == 0 {
		log.Println("No published products found, skipping cart/wishlist seeding")
		return nil
	}

	// Debug: Check if product_variants table has data
	var variantCount int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM product_variants").Scan(&variantCount)
	if err != nil {
		log.Printf("Error checking product_variants count: %v", err)
	} else {
		log.Printf("Found %d variants in product_variants table", variantCount)
	}

	// Seed cart items for demo sessions
	cartItems := []struct {
		sessionID string
		productID uuid.UUID
		quantity  int
	}{
		{"demo-session-1", productIDs[0], 1},
		{"demo-session-1", productIDs[1], 2},
		{"demo-session-2", productIDs[2], 1},
	}

	for _, item := range cartItems {
		// Get a variant for this product
		var variantID int
		log.Printf("Looking for variant for product UUID: %s", item.productID.String())
		err := pool.QueryRow(ctx, `
			SELECT pv.id FROM product_variants pv
			WHERE pv.product_id = $1::uuid 
			ORDER BY pv.id 
			LIMIT 1
		`, item.productID.String()).Scan(&variantID)
		if err != nil {
			log.Printf("No variant found for product %s, error: %v", item.productID.String(), err)
			continue
		}
		log.Printf("Found variant ID: %d for product %s", variantID, item.productID.String())

		_, err = pool.Exec(ctx, `
			INSERT INTO cart (session_id, product_id, variant_id, quantity, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`, item.sessionID, item.productID, variantID, item.quantity)
		if err != nil {
			return err
		}
	}

	// Seed wishlist items for the demo customer
	for i := 0; i < len(productIDs) && i < 3; i++ {
		_, err = pool.Exec(ctx, `
			INSERT INTO wishlist (customer_id, product_id, created_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (customer_id, product_id) DO NOTHING
		`, customerID, productIDs[i])
		if err != nil {
			return err
		}
	}

	log.Printf("Created cart items for demo sessions and wishlist items for customer %s", customerID.String())
	return nil
}

func stringPtr(s string) *string {
	return &s
}
