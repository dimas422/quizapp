package auth

import (
	"context"
	"errors"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// База данных
var db = sqldb.NewDatabase("quiz", sqldb.DatabaseConfig{
    Migrations: "./migrations",
})

var jwtSecret = []byte("super-secret-key-change-in-prod")

// ===== ТИПЫ =====

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"` // "admin" или "user"
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

type UserData struct {
	UserID string
	Role   string
	Email  string
}

// ===== РЕГИСТРАЦИЯ =====

//encore:api public method=POST path=/auth/register
func Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email и пароль обязательны")
	}
	if req.Role != "admin" && req.Role != "user" {
		return nil, errors.New("роль должна быть admin или user")
	}

	// Хэшируем пароль (как hashlib в Python)
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Сохраняем в БД
	var userID string
	err = db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id`,
		req.Email, string(hash), req.Role,
	).Scan(&userID)
	if err != nil {
		return nil, errors.New("пользователь с таким email уже существует")
	}

	// Создаём JWT токен
	token, err := createToken(userID, req.Email, req.Role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, Role: req.Role}, nil
}

// ===== ВХОД =====

//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, errors.New("email и пароль обязательны")
	}

	// Ищем пользователя в БД
	var userID, passwordHash, role string
	err := db.QueryRow(ctx,
		`SELECT id, password_hash, role FROM users WHERE email = $1`,
		req.Email,
	).Scan(&userID, &passwordHash, &role)
	if err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("неверный email или пароль")
	}

	// Создаём токен
	token, err := createToken(userID, req.Email, role)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, Role: role}, nil
}

// ===== AUTH HANDLER (проверка токена для защищённых роутов) =====

//encore:authhandler
func AuthHandler(ctx context.Context, token string) (auth.UID, *UserData, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !parsed.Valid {
		return "", nil, errors.New("неверный токен")
	}

	claims := parsed.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)
	email := claims["email"].(string)
	role := claims["role"].(string)

	return auth.UID(userID), &UserData{
		UserID: userID,
		Role:   role,
		Email:  email,
	}, nil
}

// ===== ВСПОМОГАТЕЛЬНАЯ ФУНКЦИЯ =====

func createToken(userID, email, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(jwtSecret)
}