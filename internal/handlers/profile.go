package handler

import "github.com/gin-gonic/gin"

func (h *Handler) Me(c *gin.Context) {
	claims := GetClaims(c)
	writeOK(c, gin.H{
		"data": gin.H{
			"id":    claims.UserID,
			"email": claims.Email,
			"role":  claims.Role,
		},
	})
}
