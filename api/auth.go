package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

const (
	contextAuthDisabledKey = "auth_disabled"
	contextPrincipalKey    = "auth_principal"
)

type AuthPrincipal struct {
	Subject string
	Email   string
}

type AuthMiddleware struct {
	enabled  bool
	verifier *oidc.IDTokenVerifier
}

func NewAuthMiddleware(ctx context.Context) (*AuthMiddleware, error) {
	enabled := !strings.EqualFold(strings.TrimSpace(os.Getenv("AUTH_ENABLED")), "false")
	if !enabled {
		return &AuthMiddleware{enabled: false}, nil
	}

	issuer := os.Getenv("OIDC_ISSUER_URL")
	clientID := os.Getenv("OIDC_CLIENT_ID")
	if issuer == "" || clientID == "" {
		return nil, errors.New("AUTH_ENABLED is true but OIDC_ISSUER_URL or OIDC_CLIENT_ID is missing")
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
			c.Set(contextAuthDisabledKey, true)
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
		idToken, err := a.verifier.Verify(c.Request.Context(), token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims := struct {
			Subject string `json:"sub"`
			Email   string `json:"email"`
		}{}
		if err := idToken.Claims(&claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		email := strings.TrimSpace(strings.ToLower(claims.Email))
		if email == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token is missing email claim"})
			return
		}

		c.Set(contextPrincipalKey, AuthPrincipal{
			Subject: strings.TrimSpace(claims.Subject),
			Email:   email,
		})
		c.Next()
	}
}

func principalFromContext(c *gin.Context) (AuthPrincipal, bool) {
	value, ok := c.Get(contextPrincipalKey)
	if !ok {
		return AuthPrincipal{}, false
	}
	principal, ok := value.(AuthPrincipal)
	return principal, ok
}
