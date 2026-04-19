package middleware

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PasetoTokenClaims represents the claims stored in a PASETO token
type PasetoTokenClaims struct {
	UserID   string    `json:"user_id"`
	Email    string    `json:"email"`
	Role     string    `json:"role"` // admin, pentester, stakeholder
	IssuedAt time.Time `json:"iat"`
	Exp      time.Time `json:"exp"`
}

// TokenManager handles PASETO v4 token operations (manual implementation)
// PASETO v4 public tokens use Ed25519 signatures
type TokenManager struct {
	publicKey  ed25519.PublicKey
	privateKey ed25519.PrivateKey
	expiryTime time.Duration
}

// NewTokenManager creates a new TokenManager with existing keys
func NewTokenManager(publicKeyStr, privateKeyStr string, expiryTime time.Duration) (*TokenManager, error) {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	privateKey, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	return &TokenManager{
		publicKey:  ed25519.PublicKey(publicKey),
		privateKey: ed25519.PrivateKey(privateKey),
		expiryTime: expiryTime,
	}, nil
}

// GenerateKeyPair creates a new Ed25519 key pair and returns base64 encoded strings
func GenerateKeyPair() (publicKey, privateKey string, err error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	publicKey = base64.StdEncoding.EncodeToString(pub)
	privateKey = base64.StdEncoding.EncodeToString(priv)

	return publicKey, privateKey, nil
}

// GenerateToken creates a PASETO v4 public token with user claims
// Format: v4.public.base64(payload + signature).base64(footer)
func (tm *TokenManager) GenerateToken(userID, email, role string) (string, error) {
	now := time.Now().UTC()
	exp := now.Add(tm.expiryTime)

	claims := PasetoTokenClaims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		IssuedAt: now,
		Exp:      exp,
	}

	// Serialize claims to JSON
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Create PASETO v4 public message format
	// Message = "v4.public." || payload
	header := "v4.public."
	message := []byte(header + string(claimsJSON))

	// Sign the message with Ed25519
	signature := ed25519.Sign(tm.privateKey, message)

	// Combine payload and signature
	combined := append(claimsJSON, signature...)

	// Encode to base64
	encodedPayload := base64.RawURLEncoding.EncodeToString(combined)

	// Final token format: v4.public.base64payload
	token := fmt.Sprintf("v4.public.%s", encodedPayload)

	return token, nil
}

// VerifyToken parses and validates a PASETO v4 public token
func (tm *TokenManager) VerifyToken(token string) (*PasetoTokenClaims, error) {
	// Verify token format
	if !strings.HasPrefix(token, "v4.public.") {
		return nil, fmt.Errorf("invalid token format")
	}

	// Extract payload
	encodedPayload := strings.TrimPrefix(token, "v4.public.")
	combined, err := base64.RawURLEncoding.DecodeString(encodedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token payload: %w", err)
	}

	// Ed25519 signature is 64 bytes
	signatureSize := 64
	if len(combined) < signatureSize {
		return nil, fmt.Errorf("invalid token: too short")
	}

	// Split payload and signature
	claimsJSON := combined[:len(combined)-signatureSize]
	signature := combined[len(combined)-signatureSize:]

	// Recreate the message for verification
	header := "v4.public."
	message := []byte(header + string(claimsJSON))

	// Verify signature
	if !ed25519.Verify(tm.publicKey, message, signature) {
		return nil, fmt.Errorf("invalid token signature")
	}

	// Parse claims
	var claims PasetoTokenClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	// Check if token is expired
	if claims.Exp.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("token has expired")
	}

	return &claims, nil
}

// GetPublicKeyStr returns the public key as base64 encoded string
func (tm *TokenManager) GetPublicKeyStr() string {
	return base64.StdEncoding.EncodeToString(tm.publicKey)
}

// GetPrivateKeyStr returns the private key as base64 encoded string
func (tm *TokenManager) GetPrivateKeyStr() string {
	return base64.StdEncoding.EncodeToString(tm.privateKey)
}

// AuthMiddleware creates a Fiber middleware for PASETO token validation
// Extracts token from Authorization header, verifies it, and sets claims in context
func (tm *TokenManager) AuthMiddleware() func(*fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Extract Bearer token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			// No "Bearer " prefix found
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Verify token
		claims, err := tm.VerifyToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		// Store claims in context
		c.Locals("userID", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("claims", claims)

		return c.Next()
	}
}
