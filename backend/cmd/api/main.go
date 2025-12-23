package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/etreasure/backend/internal/config"
	"github.com/etreasure/backend/internal/db"
	"github.com/etreasure/backend/internal/email"
	"github.com/etreasure/backend/internal/handlers"
	"github.com/etreasure/backend/internal/middleware"
	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	_ = godotenv.Load()
	ctx := context.Background()
	cfg := config.Load()

	pool, err := db.NewPool(ctx, cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	defer pool.Close()

	// Initialize Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test Redis connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Printf("Forgot password functionality will not work without Redis")
	} else {
		log.Println("Connected to Redis successfully")
	}

	// Initialize email service
	emailService := email.NewEmailService()

	// Test SMTP connection (optional)
	if err := emailService.TestConnection(); err != nil {
		log.Printf("Warning: SMTP connection test failed: %v", err)
		log.Printf("Email functionality may not work properly")
	} else {
		log.Println("SMTP connection test successful")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Ensure upload dir exists for local media
	_ = os.MkdirAll(cfg.UploadDir, 0o755)

	r := gin.Default()

	// During local development allow any origin so mobile devices on the LAN can reach the API.
	// In production, set DEV env var to empty (or not set) to use a restrictive whitelist.
	devMode := os.Getenv("DEV") == "true"
	if devMode {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     []string{"http://localhost:4321", "http://127.0.0.1:4321", "http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:5174", "http://127.0.0.1:5174"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	} else {
		// Get allowed origins from environment or use defaults
		allowedOrigins := []string{"http://localhost:4321", "http://localhost:3000", "http://127.0.0.1:4321", "http://localhost:5174", "http://127.0.0.1:5174"}

		// Add production frontend URLs from environment
		if adminURL := os.Getenv("ADMIN_FRONTEND_URL"); adminURL != "" {
			allowedOrigins = append(allowedOrigins, adminURL)
		}
		if webURL := os.Getenv("WEB_FRONTEND_URL"); webURL != "" {
			allowedOrigins = append(allowedOrigins, webURL)
		}

		// Fallback to hardcoded admin URL for backward compatibility
		allowedOrigins = append(allowedOrigins, "https://etreasureadmin.onrender.com")

		r.Use(cors.New(cors.Config{
			AllowOrigins:     allowedOrigins,
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}))
	}

	// serve uploaded assets
	r.Static("/uploads", cfg.UploadDir)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authHandler := &handlers.AuthHandler{DB: pool, Cfg: cfg, Rd: redisClient, Email: emailService}

	// Admin auth routes
	authGroup := r.Group("/api/admin/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/verify-otp", authHandler.VerifyOTP)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
	}

	public := r.Group("/api/public")
	{
		public.POST("/login", authHandler.Login)
		public.POST("/signup", authHandler.Signup)
		public.POST("/refresh", authHandler.Refresh)
		public.POST("/forgot-password", authHandler.ForgotPassword)
		public.POST("/reset-password", authHandler.ResetPassword)
		public.GET("/me", authHandler.Me)
	}

	// Initialize R2 client for global use
	r2Client, err := storage.NewR2Client(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}
	log.Printf("R2 client initialized successfully in web server!")

	// Initialize ImageURLHelper for consistent image URL formatting
	imageHelper := storage.NewImageURLHelper(r2Client)

	protected := r.Group("/api/admin")
	// protected.Use(middleware.AuthRequired(cfg)) // Temporarily disabled for testing
	{
		protected.GET("/me", authHandler.Me)

		// Products (admin)
		products := &handlers.ProductsHandler{DB: pool, R2Client: r2Client, ImageHelper: imageHelper}
		protected.GET("/products", products.List)
		protected.POST("/products", products.Create)
		protected.GET("/products/:id", products.Get)
		protected.PUT("/products/:id", products.Update)
		protected.DELETE("/products/:id", products.Delete)
		protected.POST("/products/import", products.Import)

		// Media
		media := &handlers.MediaHandler{DB: pool, UploadDir: cfg.UploadDir, HMACSecret: cfg.UploadHMACSecret, R2Client: r2Client, Config: cfg}
		protected.POST("/media/presign", media.Presign)
		protected.PUT("/media/upload/:id", media.Upload)
		protected.POST("/media/upload", media.UploadR2)
		protected.GET("/media", media.List)
		protected.DELETE("/media/:id", media.Delete)

		// Categories
		categories := &handlers.CategoriesHandler{DB: pool, R2Client: r2Client, ImageHelper: imageHelper}
		protected.GET("/categories", categories.List)
		protected.POST("/categories", categories.Create)
		protected.PUT("/categories/:id", categories.Update)
		protected.DELETE("/categories/:id", categories.Delete)

		// Banners
		banners := &handlers.Handler{DB: pool, R2Client: r2Client, Config: cfg, ImageHelper: imageHelper}
		protected.GET("/banners", banners.ListBanners)
		protected.POST("/banners", banners.CreateBanner)
		protected.GET("/banners/:id", banners.GetBanner)
		protected.PUT("/banners/:id", banners.UpdateBanner)
		protected.DELETE("/banners/:id", banners.DeleteBanner)

		// Offers
		offers := &handlers.Handler{DB: pool}
		protected.GET("/offers", offers.ListOffers)
		protected.POST("/offers", offers.CreateOffer)
		protected.GET("/offers/:id", offers.GetOffer)
		protected.PUT("/offers/:id", offers.UpdateOffer)
		protected.DELETE("/offers/:id", offers.DeleteOffer)

		// Orders
		orders := &handlers.Handler{DB: pool}
		protected.GET("/orders", orders.ListOrders)
		protected.POST("/orders", orders.CreateOrder)
		protected.GET("/orders/:id", orders.GetOrder)
		protected.PUT("/orders/:id", orders.UpdateOrder)
		protected.DELETE("/orders/:id", orders.DeleteOrder)
		protected.GET("/orders/debug/schema", orders.DebugOrdersSchema)
		protected.GET("/orders/debug/line-items", orders.DebugLineItems)
		protected.POST("/orders/fix-prices", orders.FixOrderLineItemsPrices)
		protected.POST("/orders/fix-null-prices", orders.FixNullPrices)

		// Customers (from users table, excluding admin roles)
		customers := &handlers.Handler{DB: pool}
		protected.GET("/customers", customers.ListUserCustomers)
		protected.GET("/customers/:id/orders", customers.GetCustomerOrders)

		// Inventory
		inventory := &handlers.Handler{DB: pool}
		protected.GET("/inventory", inventory.ListInventory)
		protected.POST("/inventory", inventory.CreateInventoryItem)
		protected.GET("/inventory/:id", inventory.GetInventoryItem)
		protected.PUT("/inventory/:id", inventory.UpdateInventoryItem)
		protected.DELETE("/inventory/:id", inventory.DeleteInventoryItem)
		protected.POST("/inventory/:id/adjust", inventory.AdjustInventory)

		// Settings
		settings := &handlers.Handler{DB: pool}
		protected.GET("/settings", settings.ListSettings)
		protected.POST("/settings", settings.CreateSetting)
		protected.GET("/settings/:key", settings.GetSetting)
		protected.PUT("/settings/:key", settings.UpdateSetting)
		protected.DELETE("/settings/:key", settings.DeleteSetting)

		// Users & Roles
		users := &handlers.Handler{DB: pool}
		protected.GET("/users", users.ListUsers)
		protected.POST("/users", users.CreateUser)
		protected.GET("/users/:id", users.GetUser)
		protected.PUT("/users/:id", users.UpdateUser)
		protected.DELETE("/users/:id", users.DeleteUser)
		protected.GET("/roles", users.ListRoles)

		// Preview
		preview := &handlers.Handler{DB: pool}
		protected.POST("/preview", preview.Preview)

		// Content Management
		content := &handlers.Handler{DB: pool}
		protected.GET("/content/pages", content.ListContentPages)
		protected.POST("/content/pages", content.CreateContentPage)
		protected.GET("/content/pages/:slug", content.GetContentPageAdmin) // Use admin handler
		protected.PUT("/content/pages/:id", content.CreateContentPage)     // Update uses same handler
		protected.DELETE("/content/pages/:id", content.DeleteContentPage)

		protected.GET("/content/faqs", content.ListFAQs)
		protected.POST("/content/faqs", content.CreateFAQ)
		protected.PUT("/content/faqs/:id", content.CreateFAQ) // Update uses same handler
		protected.DELETE("/content/faqs/:id", content.DeleteFAQ)
	}

	// Public settings (no auth)
	r.GET("/api/public/settings", (&handlers.Handler{DB: pool}).GetPublicSettings)

	// Public offers endpoint
	r.GET("/api/public/offers", (&handlers.Handler{DB: pool}).ListOffers)

	// Public categories endpoint
	publicCategories := &handlers.CategoriesHandler{DB: pool, R2Client: r2Client, ImageHelper: imageHelper}
	r.GET("/api/public/categories", publicCategories.PublicList)

	// Public banners endpoint for storefront
	publicBanners := &handlers.Handler{DB: pool, R2Client: r2Client, Config: cfg, ImageHelper: imageHelper}
	r.GET("/api/public/banners", publicBanners.ListPublicBanners)

	// Public content endpoints
	publicContent := &handlers.Handler{DB: pool}
	r.GET("/api/public/content/pages/:slug", publicContent.GetContentPage)
	r.GET("/api/public/content/faqs", publicContent.ListFAQs)

	// Public newsletter subscription endpoint
	r.POST("/api/public/newsletter/subscribe", publicBanners.CreateNewsletterSubscriber)

	// Public product endpoints for storefront
	publicProducts := &handlers.ProductsHandler{DB: pool, R2Client: r2Client, ImageHelper: imageHelper}
	r.GET("/api/products", publicProducts.PublicList)
	r.GET("/api/products/:id", publicProducts.PublicGet)
	r.POST("/api/products/search", publicProducts.Search)
	r.GET("/api/products/:id/related", publicProducts.Related)

	// Authentication endpoints
	r.POST("/api/auth/login", authHandler.Login)
	r.POST("/api/auth/signup", authHandler.Signup)
	r.POST("/api/auth/refresh", authHandler.Refresh)
	r.POST("/api/auth/logout", authHandler.Logout)
	r.GET("/api/auth/me", authHandler.Me)
	r.POST("/api/auth/forgot-password", authHandler.ForgotPassword)
	r.POST("/api/auth/verify-otp", authHandler.VerifyOTP)
	r.POST("/api/auth/reset-password", authHandler.ResetPassword)

	// Temporary endpoint to fix orders table schema
	r.POST("/admin/fix-orders-schema", func(c *gin.Context) {
		ctx := context.Background()

		sqlScript := `
-- Add missing columns to orders table
ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS customer_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS customer_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS customer_phone VARCHAR(20),
ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50) DEFAULT 'razorpay',
ADD COLUMN IF NOT EXISTS shipping_status VARCHAR(20) DEFAULT 'just arrived',
ADD COLUMN IF NOT EXISTS razorpay_order_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS razorpay_payment_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS razorpay_signature VARCHAR(255),
ADD COLUMN IF NOT EXISTS shipping_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS shipping_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS shipping_phone VARCHAR(20),
ADD COLUMN IF NOT EXISTS shipping_address_line1 TEXT,
ADD COLUMN IF NOT EXISTS shipping_address_line2 TEXT,
ADD COLUMN IF NOT EXISTS shipping_city VARCHAR(100),
ADD COLUMN IF NOT EXISTS shipping_state VARCHAR(100),
ADD COLUMN IF NOT EXISTS shipping_country VARCHAR(100) DEFAULT 'India',
ADD COLUMN IF NOT EXISTS shipping_pin_code VARCHAR(20),
ADD COLUMN IF NOT EXISTS billing_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS billing_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS billing_phone VARCHAR(20),
ADD COLUMN IF NOT EXISTS billing_address_line1 TEXT,
ADD COLUMN IF NOT EXISTS billing_address_line2 TEXT,
ADD COLUMN IF NOT EXISTS billing_city VARCHAR(100),
ADD COLUMN IF NOT EXISTS billing_state VARCHAR(100),
ADD COLUMN IF NOT EXISTS billing_country VARCHAR(100) DEFAULT 'India',
ADD COLUMN IF NOT EXISTS billing_pin_code VARCHAR(20),
ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(255),
ADD COLUMN IF NOT EXISTS tracking_provider VARCHAR(100),
ADD COLUMN IF NOT EXISTS estimated_delivery DATE,
ADD COLUMN IF NOT EXISTS subtotal DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS tax_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS shipping_amount DECIMAL(10,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10,2) DEFAULT 0;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Update existing orders to have default values if needed
UPDATE orders SET 
    customer_name = COALESCE(customer_name, 'Guest Customer'),
    customer_email = COALESCE(customer_email, 'guest@example.com'),
    customer_phone = COALESCE(customer_phone, '0000000000'),
    subtotal = COALESCE(subtotal, 0),
    tax_amount = COALESCE(tax_amount, 0),
    shipping_amount = COALESCE(shipping_amount, 0),
    discount_amount = COALESCE(discount_amount, 0),
    total_price = COALESCE(total_price, 0)
WHERE customer_name IS NULL OR customer_email IS NULL OR customer_phone IS NULL OR subtotal IS NULL OR tax_amount IS NULL OR shipping_amount IS NULL OR discount_amount IS NULL;
`

		_, err := pool.Exec(ctx, sqlScript)
		if err != nil {
			log.Printf("Failed to fix orders schema: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fix orders schema", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Orders schema fixed successfully!"})
	})

	// Temporary endpoint to set some orders as just arrived for testing
	r.POST("/admin/set-just-arrived", func(c *gin.Context) {
		ctx := context.Background()

		// Update orders to have shipping_status = 'just arrived'
		_, err := pool.Exec(ctx, `
			UPDATE orders 
			SET shipping_status = 'just arrived', updated_at = NOW()
			WHERE shipping_status IS NULL OR shipping_status = ''
		`)
		if err != nil {
			log.Printf("Failed to set just arrived status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set just arrived status", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Set 5 orders as just arrived successfully!"})
	})

	// Temporary endpoint to create missing tables
	r.POST("/admin/create-tables", func(c *gin.Context) {
		ctx := context.Background()

		sqlScript := `
-- Create cart table
CREATE TABLE IF NOT EXISTS cart (
    id SERIAL PRIMARY KEY,
    session_id TEXT, -- For guest users
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE, -- For authenticated users
    product_id UUID REFERENCES products(uuid_id) ON DELETE CASCADE,
    variant_id INT REFERENCES product_variants(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Ensure either session_id or user_id is present
    CONSTRAINT cart_user_or_session CHECK (
        (session_id IS NOT NULL AND user_id IS NULL) OR 
        (session_id IS NULL AND user_id IS NOT NULL) OR
        (session_id IS NULL AND user_id IS NULL) -- Allow empty for admin operations
    )
);

-- Create wishlist table
CREATE TABLE IF NOT EXISTS wishlist (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(uuid_id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, product_id)
);

-- Create indexes for cart
CREATE INDEX IF NOT EXISTS idx_cart_session ON cart(session_id);
CREATE INDEX IF NOT EXISTS idx_cart_user ON cart(user_id);
CREATE INDEX IF NOT EXISTS idx_cart_product ON cart(product_id);

-- Create indexes for wishlist
CREATE INDEX IF NOT EXISTS idx_wishlist_user ON wishlist(user_id);
CREATE INDEX IF NOT EXISTS idx_wishlist_product ON wishlist(product_id);
`

		_, err := pool.Exec(ctx, sqlScript)
		if err != nil {
			log.Printf("Failed to create tables: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tables", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart and wishlist tables created successfully!"})
	})

	// Public Cart and Wishlist endpoints (no authentication required)
	cartHandler := &handlers.CartHandler{DB: pool, ImageHelper: imageHelper}
	r.POST("/api/cart/add", cartHandler.AddToCart)
	r.GET("/api/cart", cartHandler.GetCart)
	r.DELETE("/api/cart/:id", cartHandler.RemoveFromCart)
	r.POST("/api/cart/clear", cartHandler.ClearCart)

	wishlistHandler := &handlers.WishlistHandler{DB: pool, ImageHelper: imageHelper}
	r.POST("/api/wishlist/toggle", wishlistHandler.ToggleWishlist)
	r.GET("/api/wishlist", wishlistHandler.GetWishlist)
	r.DELETE("/api/wishlist/:id", wishlistHandler.RemoveFromWishlist)

	// Public search endpoints (fast, cached, rate-limited)
	searchHandler := handlers.NewSearchHandler(pool)
	r.GET("/api/search", searchHandler.Search)
	r.GET("/api/search/suggest", searchHandler.Suggest)
	r.GET("/api/search/facets", searchHandler.Facets)
	r.GET("/api/search/health", searchHandler.Health)

	// Admin search endpoints
	protected.POST("/search/reindex", searchHandler.Reindex)

	r.GET("/api/search/fix", func(c *gin.Context) {
		ctx := c.Request.Context()

		// Enable required PostgreSQL extensions
		_, err := pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS pg_trgm`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable pg_trgm extension: " + err.Error()})
			return
		}

		_, err = pool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS unaccent`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable unaccent extension: " + err.Error()})
			return
		}

		// Ensure search_vector column exists
		_, err = pool.Exec(ctx, `ALTER TABLE products ADD COLUMN IF NOT EXISTS search_vector tsvector`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add search_vector column: " + err.Error()})
			return
		}

		// Update search_vector for all products (including those that already have it)
		result, err := pool.Exec(ctx, `
			UPDATE products
			SET search_vector = (
				setweight(to_tsvector('simple', coalesce(unaccent(title),'')), 'A') ||
				setweight(to_tsvector('simple', coalesce(unaccent(description),'')), 'C')
			)
		`)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update search_vector: " + err.Error()})
			return
		}

		// Create indexes for search
		_, err = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_products_search_vector ON products USING GIN(search_vector)`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create search_vector index: " + err.Error()})
			return
		}

		_, err = pool.Exec(ctx, `CREATE INDEX IF NOT EXISTS idx_products_title_trgm ON products USING GIN (lower(title) gin_trgm_ops)`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create title trigram index: " + err.Error()})
			return
		}

		updatedCount := result.RowsAffected()
		c.JSON(http.StatusOK, gin.H{
			"message":            "Search setup completed successfully",
			"products_updated":   updatedCount,
			"extensions_enabled": []string{"pg_trgm", "unaccent"},
			"indexes_created":    []string{"idx_products_search_vector", "idx_products_title_trgm"},
		})
	})

	// Payment endpoints (Razorpay)
	razorpay := &handlers.RazorpayHandler{DB: pool, Cfg: cfg}
	r.POST("/api/orders/create-payment", razorpay.CreatePayment)
	r.POST("/api/orders/verify-payment", razorpay.VerifyPayment)

	// Authenticated user orders
	userOrders := r.Group("/api/orders")
	userOrders.Use(middleware.AuthRequired(cfg))
	{
		userOrders.GET("/my", (&handlers.Handler{DB: pool}).ListMyOrders)
	}

	// GraphQL endpoint
	graphqlHandler := handlers.NewGraphQLHandler(pool)
	r.POST("/graphql", func(c *gin.Context) {
		var requestBody struct {
			Query string `json:"query"`
		}

		if err := c.ShouldBindJSON(&requestBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		result, err := graphqlHandler.ExecuteQuery(requestBody.Query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
	})

	// Public media proxy (serves R2 images locally) - after R2 client is initialized
	r.GET("/api/public/media/:key", func(c *gin.Context) {
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "key is required"})
			return
		}

		// Decode the key (replace underscores back to slashes)
		originalKey := key
		key = strings.ReplaceAll(key, "_", "/")
		_ = originalKey

		// Try to get the object from R2
		result, err := r2Client.GetObject(c.Request.Context(), key)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
			return
		}
		defer result.Body.Close()

		// Set content type
		c.Header("Content-Type", *result.ContentType)
		c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year

		// Stream the image directly
		_, err = io.Copy(c.Writer, result.Body)
		if err != nil {
			log.Printf("Media proxy: Failed to stream image: %v", err)
			return
		}
	})

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
