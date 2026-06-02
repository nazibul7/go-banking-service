package utils

import (
	"banking-app/internal/model"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateAccessToken(userID int, email string, role model.Role, expires_at time.Duration, secret string) (string, error) {
	expiresAt := time.Now().Add(expires_at)
	claim := model.Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: model.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	signed, err := token.SignedString([]byte(secret))
	return signed, err
}

func GenerateRefreshToken() (string, time.Time, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", time.Time{}, err
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	return token, expiresAt, nil
}

func VerifyToken(
	tokenString string,
	secret string,
	tokenType model.TokenType,
) (*model.Claims, error) {

	// Empty struct where JWT library will store parsed claims.
	//
	// After successful verification:
	// claim.UserID
	// claim.Role
	// claim.ExpiresAt
	// etc...
	//
	// will be automatically populated.
	claim := model.Claims{}

	// ParseWithClaims does MANY things internally:
	//
	// 1. Split token into:
	//    header.payload.signature
	//
	// 2. Decode header + payload
	//
	// 3. Read algorithm from token header
	//
	// 4. Call our callback function below
	//
	// 5. Get secret key from callback
	//
	// 6. Recompute expected signature using:
	//    HMAC(header.payload, secret)
	//
	// 7. Compare expected signature
	//    with incoming signature
	//
	// 8. Validate standard claims:
	//    exp, iat, nbf, etc...
	//
	// 9. Populate `claim` struct
	token, err := jwt.ParseWithClaims(

		// Incoming JWT token from client
		tokenString,

		// Struct where decoded claims will be stored
		&claim,

		// Callback function.
		//
		// JWT library pauses parsing here and asks:
		//
		// "What key should I use to verify this token?"
		func(t *jwt.Token) (any, error) {

			// SECURITY CHECK
			//
			// Check whether incoming token claims
			// to use expected algorithm.
			//
			// Example valid:
			// HS256
			//
			// Example invalid:
			// RS256
			// none
			//
			// This prevents algorithm confusion attacks.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			// Return server secret key.
			//
			// JWT library will NOW use this secret
			// internally to verify token signature.
			//
			// Conceptually:
			//
			// expectedSignature =
			//    HMAC(header.payload, secret)
			//
			// compare:
			// expectedSignature == incomingSignature
			// here we are returning secret so that invisible verify function can verify.
			return []byte(secret), nil
		},
	)

	// Verification failed:
	//
	// - invalid signature
	// - expired token
	// - malformed token
	// - wrong secret
	// - invalid claims
	if err != nil {
		return nil, err
	}

	// Extra safety check.
	//
	// token.Valid becomes true ONLY IF:
	//
	// - signature valid
	// - token not expired
	// - claims valid
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claim.TokenType != tokenType {
		return nil, errors.New("invalid token type")
	}

	// At this point:
	//
	// Token is cryptographically verified.
	//
	// claim struct now contains:
	//
	// claim.UserID
	// claim.Role
	// claim.ExpiresAt
	// etc...
	return &claim, nil
}

// HashToken hashes refresh/access tokens before storing in DB.
//
// Why hash tokens?
//
// If database leaks, attacker should NOT get usable refresh tokens.
//
// We NEVER store raw refresh tokens in database.
//
// Flow:
//
//	raw refresh token
//	        ↓
//	    SHA256 hash
//	        ↓
//	store hashed value in DB
//
// Later during verification:
//
//	client sends raw token
//	        ↓
//	hash incoming token again
//	        ↓
//	compare with stored hash
//
// Important:
//
// SHA256 returns binary data ([32]byte),
// so we convert it into readable hexadecimal string
// for easier DB storage and comparison.
func HashToken(token string) string {

	// Create SHA256 hash of token.
	//
	// Example:
	// "abc" -> binary cryptographic digest
	hash := sha256.Sum256([]byte(token))

	// Convert binary hash into readable hex string.
	//
	// Example:
	// [186 120 22 ...]
	// ->
	// "ba7816bf8f01..."
	//
	// hash[:] converts:
	// [32]byte -> []byte
	return hex.EncodeToString(hash[:])
}
