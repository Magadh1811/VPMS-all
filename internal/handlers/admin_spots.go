package handler

import (
	"database/sql"
    "net/http"

	"github.com/gin-gonic/gin"
)

type createSpotReq struct {
	LotID   string `json:"lotId" binding:"required"`
	LevelID string `json:"levelId" binding:"required"`
	Number  string `json:"number" binding:"required"`
}

func (h *Handler) CreateSpot(c *gin.Context) {
    var req createSpotReq
    if err := c.ShouldBindJSON(&req); err != nil {
        writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
        return
    }
    var id string
    err := h.DB.QueryRow(`
        INSERT INTO parking_spots (lot_id, level_id, number, status)
        VALUES ($1, $2, $3, 'AVAILABLE')
        RETURNING id
    `, req.LotID, req.LevelID, req.Number).Scan(&id)
    if err != nil {
        writeError(c, http.StatusBadRequest, "CREATE_SPOT_FAILED", "could not create spot", err.Error())
        return
    }
    c.JSON(http.StatusCreated, gin.H{"data": gin.H{
        "id": id, "lotId": req.LotID, "levelId": req.LevelID, "number": req.Number, "status": "AVAILABLE",
    }})
}

func (h *Handler) DeleteSpot(c *gin.Context) {
    id := c.Param("id")
    if id == "" {
        writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "id is required", nil)
        return
    }
    // disallow delete if spot is occupied
    var status string
    err := h.DB.QueryRow(`SELECT status FROM parking_spots WHERE id = $1`, id).Scan(&status)
    if err == sql.ErrNoRows {
        writeError(c, http.StatusNotFound, "SPOT_NOT_FOUND", "spot not found", nil)
        return
    } else if err != nil {
        writeError(c, http.StatusInternalServerError, "FETCH_SPOT_FAILED", "failed to fetch spot", err.Error())
        return
    }
    if status == "OCCUPIED" {
        writeError(c, http.StatusConflict, "SPOT_OCCUPIED", "cannot delete an occupied spot", nil)
        return
    }
    res, err := h.DB.Exec(`DELETE FROM parking_spots WHERE id = $1`, id)
    if err != nil {
        writeError(c, http.StatusInternalServerError, "DELETE_SPOT_FAILED", "failed to delete spot", err.Error())
        return
    }
    n, _ := res.RowsAffected()
    if n == 0 {
        writeError(c, http.StatusNotFound, "SPOT_NOT_FOUND", "spot not found", nil)
        return
    }
    writeOK(c, gin.H{"data": gin.H{"id": id, "deleted": true}})
}

