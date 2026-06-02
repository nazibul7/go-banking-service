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
}

func NewAuthService(authStore *store.AuthStore, refreshTokenStore *store.RefreshTokenStore) *AuthService {
	return &AuthService{
		authStore:         authStore,
		refreshTokenStore: refreshTokenStore,
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

	user, err := s.authStore.CreateUser(ctx, req.Email, hashPassword)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(user.ID, user.Email, user.Role, "")
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := utils.GenerateRefreshToken(user.ID, user.Email, model.RoleUser, "")
	if err != nil {
		return nil, err
	}
	hashRefreshToken := utils.HashToken(refreshToken)
	err = s.refreshTokenStore.SaveToken(ctx, user.ID, hashRefreshToken, expiresAt)
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

	accessToken, err := utils.GenerateAccessToken(existingUser.ID, existingUser.Email, existingUser.Role, "")
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := utils.GenerateRefreshToken(existingUser.ID, existingUser.Email, model.RoleUser, "")
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
	claims, err := utils.VerifyToken(refreshToken, "", model.TokenTypeRefresh)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

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

	err = s.refreshTokenStore.RevokeToken(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	accessToken, err := utils.GenerateAccessToken(claims.UserID, claims.Email, claims.Role, "")
	if err != nil {
		return nil, err
	}

	newRefreshToken, expiresAt, err := utils.GenerateRefreshToken(claims.UserID, claims.Email, claims.Role, "")
	if err != nil {
		return nil, err
	}
	hashRefreshToken := utils.HashToken(newRefreshToken)
	err = s.refreshTokenStore.SaveToken(ctx, claims.UserID, hashRefreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	user, err := s.authStore.GetUserByEmail(ctx, claims.Email)
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
