package service

import (
	"banking-app/internal/model"
	"banking-app/internal/store"
	"banking-app/internal/utils"
	"context"
	"database/sql"
	"errors"
	"time"
)

type AuthService struct {
	authStore         *store.AuthStore
	refreshTokenStore *store.RefreshTokenStore
	txStore           *store.TxStore
}

func NewAuthService(authStore *store.AuthStore, refreshTokenStore *store.RefreshTokenStore, txStore *store.TxStore) *AuthService {
	return &AuthService{
		authStore:         authStore,
		refreshTokenStore: refreshTokenStore,
		txStore:           txStore,
	}
}

func (s *AuthService) Signup(ctx context.Context, req model.SignupRequest) (*model.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("invalid request")
	}
	_, err := s.authStore.GetUserByEmail(ctx, req.Email)
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
	user, err := s.txStore.RegisterTx(ctx, req.Email, hashPassword, hashRefreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, 15*time.Minute, "")
	if err != nil {
		return nil, err
	}

	authResponse := &model.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return authResponse, nil
}

func (s *AuthService) Signin(ctx context.Context, req model.SigninRequest) (*model.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("invalid request")
	}
	existingUser, err := s.authStore.GetUserByEmail(ctx, req.Email)
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
	err = s.refreshTokenStore.SaveToken(ctx, existingUser.ID, hashRefreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	authResponse := &model.AuthResponse{
		User:         *existingUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return authResponse, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*model.AuthResponse, error) {
	tokenHash := utils.HashToken(refreshToken)
	token, err := s.refreshTokenStore.FindToken(ctx, tokenHash)

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
	err = s.txStore.RotateTokenTx(ctx, token.UserID, tokenHash, hashRefreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	user, err := s.authStore.GetUserByEmail(ctx, token.Email)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		User:         *user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string, userID int) error {
	tokenHash := utils.HashToken(refreshToken)
	token, err := s.refreshTokenStore.FindToken(ctx, tokenHash)
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

	if err = s.refreshTokenStore.RevokeToken(ctx, tokenHash); err != nil {
		return err
	}
	return nil
}
