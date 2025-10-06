package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type bookReq struct {
	SpotID    string `json:"spotId" binding:"required"`
	VehicleID string `json:"vehicleId" binding:"required"`
}

func (h *Handler) BookSpot(c *gin.Context) {
	var req bookReq
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
		return
	}
	claims := GetClaims(c)

	tx, err := h.DB.Begin()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "TX_BEGIN_FAILED", "could not start transaction", nil)
		return
	}
	defer tx.Rollback()

	// 1) lock spot row and ensure AVAILABLE
	var status string
	err = tx.QueryRow(`SELECT status FROM parking_spots WHERE id = $1 FOR UPDATE`, req.SpotID).Scan(&status)
	if err == sql.ErrNoRows {
		writeError(c, http.StatusBadRequest, "SPOT_NOT_FOUND", "spot does not exist", nil)
		return
	} else if err != nil {
		writeError(c, http.StatusInternalServerError, "SPOT_CHECK_FAILED", "failed to check spot", err.Error())
		return
	}
	if status != "AVAILABLE" {
		writeError(c, http.StatusConflict, "SPOT_NOT_AVAILABLE", "spot is not available", nil)
		return
	}

	// 2) ensure vehicle belongs to user
	var vcount int
	err = tx.QueryRow(`SELECT COUNT(1) FROM vehicles WHERE id = $1 AND user_id = $2`, req.VehicleID, claims.UserID).Scan(&vcount)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "VEHICLE_CHECK_FAILED", "failed to check vehicle", err.Error())
		return
	}
	if vcount == 0 {
		writeError(c, http.StatusForbidden, "VEHICLE_NOT_OWNED", "vehicle does not belong to user", nil)
		return
	}

	// 3) insert booking and get DB start_time
	var bookingID string
	var start time.Time
	err = tx.QueryRow(`
		INSERT INTO bookings (user_id, vehicle_id, spot_id)
		VALUES ($1, $2, $3)
		RETURNING id, start_time
	`, claims.UserID, req.VehicleID, req.SpotID).Scan(&bookingID, &start)
	if err != nil {
		writeError(c, http.StatusConflict, "BOOKING_CONFLICT", "active booking exists for spot or vehicle", err.Error())
		return
	}

	// 4) mark spot OCCUPIED
	if _, err = tx.Exec(`UPDATE parking_spots SET status = 'OCCUPIED' WHERE id = $1`, req.SpotID); err != nil {
		writeError(c, http.StatusInternalServerError, "SPOT_UPDATE_FAILED", "failed to mark spot occupied", err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(c, http.StatusInternalServerError, "TX_COMMIT_FAILED", "failed to commit booking", err.Error())
		return
	}

	writeOK(c, gin.H{"data": gin.H{
		"bookingId": bookingID,
		"userId":    claims.UserID,
		"vehicleId": req.VehicleID,
		"spotId":    req.SpotID,
		"status":    "ACTIVE",
		"startTime": toIST(start),
	}})
}

func (h *Handler) Release(c *gin.Context) {
	spotID := c.Param("spotId")
	if spotID == "" {
		writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "spotId is required", nil)
		return
	}
	claims := GetClaims(c)

	tx, err := h.DB.Begin()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "TX_BEGIN_FAILED", "could not start transaction", nil)
		return
	}
	defer tx.Rollback()

	// close active booking for this spot owned by this user and get DB end_time
	var end time.Time
	err = tx.QueryRow(`
		UPDATE bookings
		SET end_time = now()
		WHERE spot_id = $1 AND end_time IS NULL AND user_id = $2
		RETURNING end_time
	`, spotID, claims.UserID).Scan(&end)
	if err == sql.ErrNoRows {
		writeError(c, http.StatusConflict, "NO_ACTIVE_BOOKING", "no active booking found for this spot and user", nil)
		return
	} else if err != nil {
		writeError(c, http.StatusInternalServerError, "BOOKING_CLOSE_FAILED", "failed to close booking", err.Error())
		return
	}

	// mark spot AVAILABLE
	if _, err = tx.Exec(`UPDATE parking_spots SET status = 'AVAILABLE' WHERE id = $1`, spotID); err != nil {
		writeError(c, http.StatusInternalServerError, "SPOT_UPDATE_FAILED", "failed to mark spot available", err.Error())
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(c, http.StatusInternalServerError, "TX_COMMIT_FAILED", "failed to commit release", err.Error())
		return
	}

	writeOK(c, gin.H{"data": gin.H{
		"spotId":   spotID,
		"userId":   claims.UserID,
		"endTime":  toIST(end),
		"released": true,
	}})
}

func (h *Handler) UserHistory(c *gin.Context) {
	claims := GetClaims(c)

	rows, err := h.DB.Query(`
		SELECT id, spot_id, vehicle_id, start_time, end_time
		FROM bookings
		WHERE user_id = $1
		ORDER BY start_time DESC
		LIMIT 100
	`, claims.UserID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "HISTORY_FETCH_FAILED", "failed to fetch history", err.Error())
		return
	}
	defer rows.Close()

	items := make([]gin.H, 0, 20)
	for rows.Next() {
		var id, spotID, vehicleID string
		var start time.Time
		var end sql.NullTime
		if err := rows.Scan(&id, &spotID, &vehicleID, &start, &end); err != nil {
			writeError(c, http.StatusInternalServerError, "SCAN_ERROR", "failed to read row", err.Error())
			return
		}
		var endVal interface{}
		if end.Valid {
			endVal = toIST(end.Time)
		} else {
			endVal = nil
		}
		items = append(items, gin.H{
			"bookingId": id,
			"spotId":    spotID,
			"vehicleId": vehicleID,
			"startTime": toIST(start),
			"endTime":   endVal,
			"status":    map[bool]string{true: "COMPLETED", false: "ACTIVE"}[end.Valid],
		})
	}
	writeOK(c, gin.H{"items": items})
}
