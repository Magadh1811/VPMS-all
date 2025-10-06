package handler

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
)

func (h *Handler) Occupancy(c *gin.Context) {
    // summary
    var total, available, occupied int
    _ = h.DB.QueryRow(`SELECT COUNT(*), 
                              COUNT(*) FILTER (WHERE status='AVAILABLE'),
                              COUNT(*) FILTER (WHERE status='OCCUPIED')
                       FROM parking_spots`).Scan(&total, &available, &occupied)

    // active list
    rows, err := h.DB.Query(`
        SELECT b.id, b.user_id, b.vehicle_id, b.spot_id, b.start_time
        FROM bookings b
        WHERE b.end_time IS NULL
        ORDER BY b.start_time DESC
    `)
    if err != nil {
        writeError(c, http.StatusInternalServerError, "OCCUPANCY_FETCH_FAILED", "failed to fetch active bookings", err.Error())
        return
    }
    defer rows.Close()

    active := make([]gin.H, 0, 20)
    for rows.Next() {
        var id, uid, vid, sid string
        var start time.Time
        if err := rows.Scan(&id, &uid, &vid, &sid, &start); err != nil {
            writeError(c, http.StatusInternalServerError, "SCAN_ERROR", "failed to read row", err.Error())
            return
        }
        active = append(active, gin.H{"bookingId": id, "userId": uid, "vehicleId": vid, "spotId": sid, "startTime": start})
    }

    writeOK(c, gin.H{
        "summary": gin.H{
            "totalSpots":    total,
            "available":     available,
            "occupied":      occupied,
            "occupancyRate": func() float64 { if total == 0 { return 0 }; return float64(occupied) / float64(total) }(),
        },
        "active": active,
    })
}
