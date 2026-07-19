package store

import (
	"banking-app/internal/model"
	"context"
	"time"
)

type RefreshTokenStore struct{}

func NewRefreshTokenStore() *RefreshTokenStore {
	return &RefreshTokenStore{}
}

func (s *RefreshTokenStore) SaveToken(ctx context.Context, db DBTX, userID int, hashToken string, expires time.Time) error {
	query := `INSERT INTO refresh_tokens(user_id,token_hash, expires_at) VALUES($1, $2, $3)`
	_, err := db.ExecContext(ctx, query, userID, hashToken, expires)
	return err
}
func (s *RefreshTokenStore) DeleteToken(ctx context.Context, db DBTX, tokenHash string) error {
	query := `DELETE FROM refresh_tokens WHERE token_hash = $1`
	_, err := db.ExecContext(ctx, query, tokenHash)
	return err
}
func (s *RefreshTokenStore) FindToken(
	ctx context.Context,
	db DBTX,
	tokenHash string,
) (*model.RefreshToken, error) {

	query := `
	SELECT
		rt.id,
		rt.user_id,
		rt.token_hash,
		rt.expires_at,
		rt.revoked,
		rt.created_at,
		u.email,
		u.role
	FROM refresh_tokens rt
	JOIN users u
	ON rt.user_id=u.id
	WHERE rt.token_hash = $1
	`

	var token model.RefreshToken

	err := db.QueryRowContext(
		ctx,
		query,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Revoked,
		&token.CreatedAt,
		&token.Email,
		&token.Role,
	)

	if err != nil {
		return nil, err
	}

	return &token, nil
}
func (s *RefreshTokenStore) RevokeToken(
	ctx context.Context,
	db DBTX,
	tokenHash string,
) error {

	query := `
	UPDATE refresh_tokens
	SET revoked = TRUE
	WHERE token_hash = $1
	`

	_, err := db.ExecContext(
		ctx,
		query,
		tokenHash,
	)

	return err
}
