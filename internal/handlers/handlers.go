package handler

import (
	"database/sql"
	"net/http"
	"time"

	"Backend-Go/internal/config"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	DB  *sql.DB
	Cfg *config.Config
}

func New(db *sql.DB, cfg *config.Config) *Handler {
	return &Handler{DB: db, Cfg: cfg}
}

type ErrorResponse struct {
	Error struct {
		Code    string      `json:"code"`
		Message string      `json:"message"`
		Details interface{} `json:"details,omitempty"`
	} `json:"error"`
}

func writeError(c *gin.Context, status int, code, message string, details interface{}) {
	var er ErrorResponse
	er.Error.Code = code
	er.Error.Message = message
	er.Error.Details = details
	c.JSON(status, er)
}

func writeOK(c *gin.Context, payload interface{}) { c.JSON(http.StatusOK, payload) }

type AuthClaims struct {
	UserID string
	Email  string
	Role   string
}

func GetClaims(c *gin.Context) AuthClaims {
	id, _ := c.Get("user_id")
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	return AuthClaims{
		UserID: asString(id),
		Email:  asString(email),
		Role:   asString(role),
	}
}

func asString(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// --- IST helpers ---
var istLoc *time.Location

func init() {
	// Load once; fallback to fixed zone if tz database isn't available
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*3600+30*60) // UTC+5:30
	}
	istLoc = loc
}

func nowIST() time.Time { return time.Now().In(istLoc) }
func toIST(t time.Time) time.Time { return t.In(istLoc) }
