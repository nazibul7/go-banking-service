package service

import (
	"banking-app/internal/dto"
	"banking-app/internal/store"
	"banking-app/internal/utils"
	"context"
	"database/sql"
	"errors"
	"time"
)

type AuthService struct {
	db                *sql.DB
	authStore         *store.AuthStore
	refreshTokenStore *store.RefreshTokenStore
}

func NewAuthService(db *sql.DB, authStore *store.AuthStore, refreshTokenStore *store.RefreshTokenStore) *AuthService {
	return &AuthService{
		db:                db,
		authStore:         authStore,
		refreshTokenStore: refreshTokenStore,
	}
}

func (s *AuthService) Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("invalid request")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = s.authStore.GetUserByEmail(ctx, tx, req.Email)
	if err == nil {
		return nil, errors.New("email already in use")
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	hashPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	hashRefreshToken := utils.HashToken(refreshToken)

	user, err := s.authStore.CreateUser(ctx, tx, req.Email, hashPassword)
	if err != nil {
		return nil, err
	}

	if err := s.refreshTokenStore.SaveToken(ctx, tx, user.ID, hashRefreshToken, expiresAt); err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, 15*time.Minute, "")
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	authResponse := &dto.AuthResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "User created successfully",
	}
	return authResponse, nil
}

func (s *AuthService) Signin(ctx context.Context, req dto.SigninRequest) (*dto.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("invalid request")
	}
	existingUser, err := s.authStore.GetUserByEmail(ctx, s.db, req.Email)
	if err != nil {
		return nil, err
	}

	err = utils.VerifyPassword(existingUser.PasswordHash, req.Password)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(existingUser.ID, existingUser.Email, existingUser.Role, 15*time.Minute, "")
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	hashRefreshToken := utils.HashToken(refreshToken)

	err = s.refreshTokenStore.SaveToken(ctx, s.db, existingUser.ID, hashRefreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	authResponse := &dto.AuthResponse{
		User: dto.UserResponse{
			ID:    existingUser.ID,
			Email: existingUser.Email,
			Role:  existingUser.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Message:      "Signed in successfully",
	}

	return authResponse, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*dto.AuthResponse, error) {
	tokenHash := utils.HashToken(refreshToken)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	token, err := s.refreshTokenStore.FindToken(ctx, tx, tokenHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("invalid refresh token")
	}
	if err != nil {
		return nil, err
	}

	if token.Revoked {
		return nil, errors.New("refresh token has been revoked")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, errors.New("refresh token has expired")
	}

	accessToken, err := utils.GenerateAccessToken(token.UserID, token.Email, token.Role, 15*time.Minute, "")
	if err != nil {
		return nil, err
	}

	newRefreshToken, expiresAt, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	hashRefreshToken := utils.HashToken(newRefreshToken)

	err = s.refreshTokenStore.RevokeToken(ctx, tx, tokenHash)
	if err != nil {
		return nil, err
	}

	err = s.refreshTokenStore.SaveToken(ctx, tx, token.UserID, hashRefreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		User: dto.UserResponse{
			ID:    token.UserID,
			Email: token.Email,
			Role:  token.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		Message:      "Token refreshed successfully",
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string, userID int) error {
	tokenHash := utils.HashToken(refreshToken)
	token, err := s.refreshTokenStore.FindToken(ctx, s.db, tokenHash)
	if errors.Is(err, sql.ErrNoRows) {
		return errors.New("invalid refresh token")
	}
	if err != nil {
		return err
	}

	if token.UserID != userID {
		return errors.New("forbidden")
	}

	if token.Revoked {
		return errors.New("refresh token has been revoked")
	}

	if time.Now().After(token.ExpiresAt) {
		return errors.New("refresh token has expired")
	}

	if err = s.refreshTokenStore.RevokeToken(ctx, s.db, tokenHash); err != nil {
		return err
	}
	return nil
}
