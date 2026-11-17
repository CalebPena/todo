package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"todo/database"
	"todo/requests"

	"github.com/gin-gonic/gin"
)

type mcpToken struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	CreatedAt  time.Time  `json:"createdAt"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	RevokedAt  *time.Time `json:"revokedAt"`
}

const mcpTokenColumns = "id, name, prefix, created_at, last_used_at, revoked_at"

func MCPTokenRoutes(group *gin.RouterGroup) {
	group.Use(func(c *gin.Context) { c.Header("Cache-Control", "no-store") })
	group.GET("", listMCPTokens)
	group.POST("", createMCPToken)
	group.POST("/:id/rotate", rotateMCPToken)
	group.DELETE("/:id", revokeMCPToken)
}

func listMCPTokens(c *gin.Context) {
	rows, err := database.GetDb(c).QueryContext(c, "SELECT "+mcpTokenColumns+" FROM mcp_tokens WHERE user_id = $1 ORDER BY created_at DESC, id DESC", GetUser(c).Id)
	if err != nil {
		requests.InternalError(c, err)
		return
	}
	defer rows.Close()
	tokens := []mcpToken{}
	for rows.Next() {
		var token mcpToken
		if err = rows.Scan(&token.ID, &token.Name, &token.Prefix, &token.CreatedAt, &token.LastUsedAt, &token.RevokedAt); err != nil {
			requests.InternalError(c, err)
			return
		}
		tokens = append(tokens, token)
	}
	if err = rows.Err(); err != nil {
		requests.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"tokens": tokens}))
}

func newMCPSecret() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	secret := "todo_mcp_" + base64.RawURLEncoding.EncodeToString(bytes)
	return secret, hashMCPSecret(secret), nil
}

func hashMCPSecret(secret string) string {
	hash := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(hash[:])
}

func createMCPToken(c *gin.Context) {
	var input struct {
		Name string `json:"name"`
	}
	if c.ShouldBindJSON(&input) != nil {
		requests.InvalidJson(c)
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 80 {
		c.JSON(400, requests.ErrorResponseWrapper("Token name must be between 1 and 80 bytes"))
		return
	}
	secret, hash, err := newMCPSecret()
	if err != nil {
		requests.InternalError(c, err)
		return
	}
	var token mcpToken
	err = database.GetDb(c).QueryRowContext(c, `INSERT INTO mcp_tokens (user_id, name, token_hash, prefix) VALUES ($1, $2, $3, $4) RETURNING `+mcpTokenColumns,
		GetUser(c).Id, input.Name, hash, secret[:17]).Scan(&token.ID, &token.Name, &token.Prefix, &token.CreatedAt, &token.LastUsedAt, &token.RevokedAt)
	if err != nil {
		requests.InternalError(c, err)
		return
	}
	c.JSON(http.StatusCreated, requests.ResponseWrapper(gin.H{"token": token, "secret": secret}))
}

func rotateMCPToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(400, requests.ErrorResponseWrapper("Invalid token ID"))
		return
	}
	secret, hash, err := newMCPSecret()
	if err != nil {
		requests.InternalError(c, err)
		return
	}
	var token mcpToken
	err = database.GetDb(c).QueryRowContext(c, `UPDATE mcp_tokens SET token_hash = $1, prefix = $2, created_at = now(), last_used_at = NULL
		WHERE id = $3 AND user_id = $4 AND revoked_at IS NULL RETURNING `+mcpTokenColumns,
		hash, secret[:17], id, GetUser(c).Id).Scan(&token.ID, &token.Name, &token.Prefix, &token.CreatedAt, &token.LastUsedAt, &token.RevokedAt)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, requests.ErrorResponseWrapper("Active token not found"))
		return
	}
	if err != nil {
		requests.InternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, requests.ResponseWrapper(gin.H{"token": token, "secret": secret}))
}

func revokeMCPToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		c.JSON(400, requests.ErrorResponseWrapper("Invalid token ID"))
		return
	}
	result, err := database.GetDb(c).ExecContext(c, "UPDATE mcp_tokens SET revoked_at = COALESCE(revoked_at, now()) WHERE id = $1 AND user_id = $2", id, GetUser(c).Id)
	if err != nil {
		requests.InternalError(c, err)
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		requests.InternalError(c, err)
		return
	}
	if count == 0 {
		c.JSON(404, requests.ErrorResponseWrapper("Token not found"))
		return
	}
	c.Status(http.StatusNoContent)
}

// MCP keys authorize only the MCP endpoint, never session or token-management routes.
func MCPAuthMiddleware(c *gin.Context) {
	parts := strings.Fields(c.GetHeader("Authorization"))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || !strings.HasPrefix(parts[1], "todo_mcp_") || len(parts[1]) != 52 {
		c.Header("WWW-Authenticate", `Bearer realm="mcp"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, requests.ErrorResponseWrapper("Invalid or revoked MCP token"))
		return
	}
	var user User
	err := database.GetDb(c).QueryRowContext(c, `UPDATE mcp_tokens SET last_used_at = now()
		WHERE token_hash = $1 AND revoked_at IS NULL RETURNING user_id`, hashMCPSecret(parts[1])).Scan(&user.Id)
	if errors.Is(err, sql.ErrNoRows) {
		c.Header("WWW-Authenticate", `Bearer realm="mcp"`)
		c.AbortWithStatusJSON(http.StatusUnauthorized, requests.ErrorResponseWrapper("Invalid or revoked MCP token"))
		return
	}
	if err != nil {
		requests.InternalError(c, err)
		c.Abort()
		return
	}
	c.Set(UserKey, user)
}
