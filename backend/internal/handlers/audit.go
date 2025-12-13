package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditHandler struct {
	DB *pgxpool.Pool
}

type AuditEntry struct {
	ID          int64     `json:"id"`
	ActorUserID *int      `json:"actor_user_id"`
	Action      string    `json:"action"`
	ObjectType  string    `json:"object_type"`
	ObjectID    string    `json:"object_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// GET /api/admin/audit?object_type=&object_id=&limit=50
func (h *AuditHandler) List(c *gin.Context) {
	ctx := context.Background()
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	objType := c.Query("object_type")
	objID := c.Query("object_id")

	var (
		rows pgx.Rows
		err  error
	)

	if objType != "" && objID != "" {
		rows, err = h.DB.Query(ctx, `SELECT id, actor_user_id, action, object_type, object_id, created_at 
            FROM audit_logs WHERE object_type=$1 AND object_id=$2 
            ORDER BY created_at DESC LIMIT $3`, objType, objID, limit)
	} else {
		rows, err = h.DB.Query(ctx, `SELECT id, actor_user_id, action, object_type, object_id, created_at 
            FROM audit_logs ORDER BY created_at DESC LIMIT $1`, limit)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	var items []AuditEntry
	for rows.Next() {
		var (
			id                 int64
			actor              sql.NullInt32
			action, otype, oid string
			created            time.Time
		)
		if err := rows.Scan(&id, &actor, &action, &otype, &oid, &created); err == nil {
			var actorPtr *int
			if actor.Valid {
				v := int(actor.Int32)
				actorPtr = &v
			}
			items = append(items, AuditEntry{ID: id, ActorUserID: actorPtr, Action: action, ObjectType: otype, ObjectID: oid, CreatedAt: created})
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
