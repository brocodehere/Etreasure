package handlers

import (
	"github.com/etreasure/backend/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB          *pgxpool.Pool
	R2Client    *storage.R2Client
	Config      interface{} // We'll use this to get R2 public URL
	ImageHelper *storage.ImageURLHelper
}
