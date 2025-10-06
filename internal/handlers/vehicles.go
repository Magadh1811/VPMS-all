package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type addVehicleReq struct {
	Plate string `json:"plate" binding:"required"`
	Type  string `json:"type" binding:"required"`
}

func (h *Handler) AddVehicle(c *gin.Context) {
    var req addVehicleReq
    if err := c.ShouldBindJSON(&req); err != nil {
        writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
        return
    }
    claims := GetClaims(c)

    var id string
    err := h.DB.QueryRow(`
        INSERT INTO vehicles (user_id, plate, type)
        VALUES ($1, $2, $3)
        RETURNING id
    `, claims.UserID, req.Plate, req.Type).Scan(&id)
    if err != nil {
        writeError(c, http.StatusBadRequest, "ADD_VEHICLE_FAILED", "could not add vehicle (maybe duplicate plate)", err.Error())
        return
    }

    writeOK(c, gin.H{"data": gin.H{
        "id":     id,
        "userId": claims.UserID,
        "plate":  req.Plate,
        "type":   req.Type,
    }})
}
