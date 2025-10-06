package handler

import (
    "database/sql"
    "net/http"
    "time"

    "golang.org/x/crypto/bcrypt"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type signupReq struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type authUser struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type authResp struct {
	Token string   `json:"token"`
	User  authUser `json:"user"`
}

func (h *Handler) Signup(c *gin.Context) {
    var req signupReq
    if err := c.ShouldBindJSON(&req); err != nil {
        writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
        return
    }

    // hash
    hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.Cfg.BcryptCost)
    if err != nil {
        writeError(c, http.StatusInternalServerError, "HASH_ERROR", "failed to hash password", nil)
        return
    }

    // insert user
    var id string
    err = h.DB.QueryRow(`
        INSERT INTO users (name, email, password_hash, role)
        VALUES ($1, $2, $3, 'user')
        RETURNING id
    `, req.Name, req.Email, string(hashed)).Scan(&id)
    if err != nil {
        writeError(c, http.StatusBadRequest, "SIGNUP_FAILED", "could not create user (maybe email exists)", err.Error())
        return
    }

    token, _ := h.signJWT(id, req.Email, "user")
    c.JSON(http.StatusCreated, authResp{
        Token: token,
        User:  authUser{ID: id, Name: req.Name, Email: req.Email, Role: "user"},
    })
}

func (h *Handler) Login(c *gin.Context) {
    var req loginReq
    if err := c.ShouldBindJSON(&req); err != nil {
        writeError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid request body", err.Error())
        return
    }

    // fetch by email (case-insensitive)
    var id, name, email, role, passwordHash string
    err := h.DB.QueryRow(`
        SELECT id, name, email, role, password_hash
        FROM users
        WHERE lower(email) = lower($1)
        LIMIT 1
    `, req.Email).Scan(&id, &name, &email, &role, &passwordHash)
    if err == sql.ErrNoRows {
        writeError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect", nil)
        return
    } else if err != nil {
        writeError(c, http.StatusInternalServerError, "LOGIN_ERROR", "failed to fetch user", err.Error())
        return
    }

    if bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)) != nil {
        writeError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "email or password is incorrect", nil)
        return
    }

    token, _ := h.signJWT(id, email, role)
    writeOK(c, authResp{
        Token: token,
        User:  authUser{ID: id, Name: name, Email: email, Role: role},
    })
}

func (h *Handler) signJWT(userID, email, role string) (string, error) {
    claims := jwt.MapClaims{
        "uid":   userID,
        "email": email,
        "role":  role,
        "exp":   time.Now().Add(24 * time.Hour).Unix(),
        "iat":   time.Now().Unix(),
    }
    t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return t.SignedString([]byte(h.Cfg.JWTSecret))
}
