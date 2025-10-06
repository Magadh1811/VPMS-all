package handler

import (
    "database/sql"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Reports(c *gin.Context) {
    var totalSessions int
    var avgMinutes sql.NullFloat64

    _ = h.DB.QueryRow(`
        SELECT COUNT(*),
               AVG(EXTRACT(EPOCH FROM (end_time - start_time))/60.0)
        FROM bookings
        WHERE end_time IS NOT NULL
    `).Scan(&totalSessions, &avgMinutes)

    var avg float64
    if avgMinutes.Valid { avg = avgMinutes.Float64 } else { avg = 0 }

    writeOK(c, gin.H{"data": gin.H{
        "totalSessions":   totalSessions,
        "avgDurationMins": avg,
    }})
}

