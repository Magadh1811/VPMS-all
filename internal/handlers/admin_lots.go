package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type createLotReq struct {
	Name string `json:"name" binding:"required"`
}

func (h *Handler) CreateLot(c *gin.Context) {
    var req createLotReq
    if err := c.ShouldBindJSON(&req); err != nil {
        writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
        return
    }
    var id string
    err := h.DB.QueryRow(`INSERT INTO parking_lots (name) VALUES ($1) RETURNING id`, req.Name).Scan(&id)
    if err != nil {
        writeError(c, http.StatusBadRequest, "CREATE_LOT_FAILED", "could not create lot (maybe duplicate name)", err.Error())
        return
    }
    c.JSON(http.StatusCreated, gin.H{"data": gin.H{"id": id, "name": req.Name}})
}
