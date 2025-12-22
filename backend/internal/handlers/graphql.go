package handlers

import (
	"context"
	"log"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GraphQLHandler struct {
	DB *pgxpool.Pool
}

func NewGraphQLHandler(db *pgxpool.Pool) *GraphQLHandler {
	return &GraphQLHandler{DB: db}
}

// GraphQL schema types
var dashboardStatsType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "DashboardStats",
		Fields: graphql.Fields{
			"totalProducts": &graphql.Field{
				Type: graphql.Int,
			},
			"totalCategories": &graphql.Field{
				Type: graphql.Int,
			},
			"totalBanners": &graphql.Field{
				Type: graphql.Int,
			},
			"activeBanners": &graphql.Field{
				Type: graphql.Int,
			},
			"totalOffers": &graphql.Field{
				Type: graphql.Int,
			},
			"activeOffers": &graphql.Field{
				Type: graphql.Int,
			},
			"recentOrders": &graphql.Field{
				Type: graphql.Int,
			},
			"totalRevenue": &graphql.Field{
				Type: graphql.Float,
			},
		},
	},
)

var orderType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Order",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"order_number": &graphql.Field{
				Type: graphql.String,
			},
			"customer_name": &graphql.Field{
				Type: graphql.String,
			},
			"customer_email": &graphql.Field{
				Type: graphql.String,
			},
			"total_price": &graphql.Field{
				Type: graphql.Float,
			},
			"currency": &graphql.Field{
				Type: graphql.String,
			},
			"status": &graphql.Field{
				Type: graphql.String,
			},
			"shipping_status": &graphql.Field{
				Type: graphql.String,
			},
			"created_at": &graphql.Field{
				Type: graphql.String,
			},
			"updated_at": &graphql.Field{
				Type: graphql.String,
			},
			"line_items": &graphql.Field{
				Type: graphql.String,
			},
		},
	},
)

var rootQuery = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"dashboardStats": &graphql.Field{
				Type: dashboardStatsType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := context.Background()
					db := p.Context.Value("db").(*pgxpool.Pool)

					// Get total products
					var totalProducts int
					err := db.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&totalProducts)
					if err != nil {
						log.Printf("Error counting products: %v", err)
						totalProducts = 0
					}

					// Get total categories
					var totalCategories int
					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM categories").Scan(&totalCategories)
					if err != nil {
						log.Printf("Error counting categories: %v", err)
						totalCategories = 0
					}

					// Get total and active banners
					var totalBanners, activeBanners int
					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM banners").Scan(&totalBanners)
					if err != nil {
						log.Printf("Error counting banners: %v", err)
						totalBanners = 0
					}
					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM banners WHERE is_active = true").Scan(&activeBanners)
					if err != nil {
						log.Printf("Error counting active banners: %v", err)
						activeBanners = 0
					}

					// Get total and active offers
					var totalOffers, activeOffers int
					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM offers").Scan(&totalOffers)
					if err != nil {
						log.Printf("Error counting offers: %v", err)
						totalOffers = 0
					}
					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM offers WHERE is_active = true").Scan(&activeOffers)
					if err != nil {
						log.Printf("Error counting active offers: %v", err)
						activeOffers = 0
					}

					// Get recent paid orders count
					var recentOrders int
					err = db.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE status = 'paid'").Scan(&recentOrders)
					if err != nil {
						log.Printf("Error counting paid orders: %v", err)
						recentOrders = 0
					}

					// Get total revenue from paid orders (already in INR)
					var totalRevenue float64
					err = db.QueryRow(ctx, "SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE status = 'paid'").Scan(&totalRevenue)
					if err != nil {
						log.Printf("Error calculating paid revenue: %v", err)
						totalRevenue = 0
					}

					return map[string]interface{}{
						"totalProducts":   totalProducts,
						"totalCategories": totalCategories,
						"totalBanners":    totalBanners,
						"activeBanners":   activeBanners,
						"totalOffers":     totalOffers,
						"activeOffers":    activeOffers,
						"recentOrders":    recentOrders,
						"totalRevenue":    totalRevenue,
					}, nil
				},
			},
			"justArrivedOrders": &graphql.Field{
				Type: graphql.NewList(orderType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					ctx := context.Background()
					db := p.Context.Value("db").(*pgxpool.Pool)

					// Get paid orders with shipping_status = 'just arrived', sorted by created_at descending (newest first)
					log.Printf("Executing query to fetch just arrived orders...")
					rows, err := db.Query(ctx, `
						SELECT o.id, o.order_number, o.customer_name, o.customer_email, o.total_price, o.currency, o.status, o.shipping_status, o.created_at, o.updated_at,
							   COALESCE(
								JSON_AGG(
									JSON_BUILD_OBJECT(
										'title', li.product_title,
										'quantity', li.quantity
									)
								) FILTER (WHERE li.id IS NOT NULL), 
								'[]'::json
							   ) as line_items
						FROM orders o
						LEFT JOIN order_line_items li ON o.id = li.order_id
						WHERE o.shipping_status = 'just arrived' AND o.status = 'paid'
						GROUP BY o.id, o.order_number, o.customer_name, o.customer_email, o.total_price, o.currency, o.status, o.shipping_status, o.created_at, o.updated_at
						ORDER BY o.created_at DESC
						LIMIT 5
					`)
					if err != nil {
						log.Printf("Error fetching just arrived orders: %v", err)
						return []interface{}{}, nil
					}
					defer rows.Close()

					log.Printf("Query executed successfully, processing rows...")

					var orders []interface{}
					orderCount := 0
					for rows.Next() {
						orderCount++
						log.Printf("Processing order #%d", orderCount)
						var order struct {
							ID             string    `json:"id"`
							OrderNumber    string    `json:"order_number"`
							CustomerName   string    `json:"customer_name"`
							Email          string    `json:"customer_email"`
							TotalPrice     float64   `json:"total_price"`
							Currency       string    `json:"currency"`
							Status         string    `json:"status"`
							ShippingStatus string    `json:"shipping_status"`
							CreatedAt      time.Time `json:"created_at"`
							UpdatedAt      time.Time `json:"updated_at"`
							LineItems      string    `json:"line_items"`
						}

						err := rows.Scan(
							&order.ID,
							&order.OrderNumber,
							&order.CustomerName,
							&order.Email,
							&order.TotalPrice,
							&order.Currency,
							&order.Status,
							&order.ShippingStatus,
							&order.CreatedAt,
							&order.UpdatedAt,
							&order.LineItems,
						)
						if err != nil {
							log.Printf("Error scanning order: %v", err)
							continue
						}

						orders = append(orders, order)
					}

					log.Printf("Finished processing. Found %d orders with shipping_status = 'just arrived'", orderCount)
					return orders, nil
				},
			},
		},
	},
)

var schema, _ = graphql.NewSchema(
	graphql.SchemaConfig{
		Query: rootQuery,
	},
)

func (h *GraphQLHandler) ExecuteQuery(query string) (interface{}, error) {
	// Create context with database connection
	ctx := context.WithValue(context.Background(), "db", h.DB)

	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
		Context:       ctx,
	})

	if len(result.Errors) > 0 {
		log.Printf("GraphQL errors: %v", result.Errors)
		return nil, result.Errors[0]
	}

	return result, nil
}
