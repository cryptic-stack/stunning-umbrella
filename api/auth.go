package main

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	enabled  bool
	verifier *oidc.IDTokenVerifier
}

func NewAuthMiddleware(ctx context.Context) (*AuthMiddleware, error) {
	enabled := strings.EqualFold(os.Getenv("AUTH_ENABLED"), "true")
	if !enabled {
		return &AuthMiddleware{enabled: false}, nil
	}

	issuer := os.Getenv("OIDC_ISSUER_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	if issuer == "" || clientID == "" {
		return &AuthMiddleware{enabled: false}, nil
	}

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}

	return &AuthMiddleware{
		enabled: true,
		verifier: provider.Verifier(&oidc.Config{
			ClientID: clientID,
		}),
	}, nil
}

func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	if a == nil || !a.enabled || a.verifier == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(header), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimSpace(header[len("Bearer "):])
		if _, err := a.verifier.Verify(c.Request.Context(), token); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}
