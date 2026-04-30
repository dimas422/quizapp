package auth

import (
"context"
"database/sql"
"errors"
"time"

"encore.app/ent"
entuser "encore.app/ent/user"
encoreauth "encore.dev/beta/auth"
"encore.dev/storage/sqldb"
"github.com/golang-jwt/jwt/v5"
"github.com/google/uuid"
"golang.org/x/crypto/bcrypt"
)

var quizDB = sqldb.Named("quiz")

var authDB = sqldb.NewDatabase("quiz", sqldb.DatabaseConfig{
Migrations: "./migrations",
})

var jwtSecret = []byte("super-secret-key-change-in-prod")

type RegisterRequest struct {
Email    string `json:"email"`
Password string `json:"password"`
Role     string `json:"role"`
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

//encore:api public method=POST path=/auth/register
func Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
if req.Email == "" || req.Password == "" {
return nil, errors.New("email и пароль обязательны")
}
if req.Role != "admin" && req.Role != "user" {
return nil, errors.New("роль должна быть admin или user")
}

hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
return nil, err
}

client, err := ent.OpenEntClient(quizDB.Stdlib())
if err != nil {
return nil, err
}
defer client.Close()

user, err := client.User.Create().
SetEmail(req.Email).
SetPasswordHash(string(hash)).
SetRole(req.Role).
SetCreatedAt(time.Now()).
Save(ctx)
if err != nil {
return nil, errors.New("пользователь с таким email уже существует")
}

token, err := createToken(user.ID.String(), req.Email, req.Role)
if err != nil {
return nil, err
}

return &AuthResponse{Token: token, Role: req.Role}, nil
}

//encore:api public method=POST path=/auth/login
func Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
if req.Email == "" || req.Password == "" {
return nil, errors.New("email и пароль обязательны")
}

client, err := ent.OpenEntClient(authDB.Stdlib())
if err != nil {
return nil, err
}
defer client.Close()

user, err := client.User.Query().
Where(entuser.Email(req.Email)).
Only(ctx)
if err != nil {
return nil, errors.New("неверный email или пароль")
}

err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
if err != nil {
return nil, errors.New("неверный email или пароль")
}

token, err := createToken(user.ID.String(), req.Email, user.Role)
if err != nil {
return nil, err
}

return &AuthResponse{Token: token, Role: user.Role}, nil
}

//encore:authhandler
func AuthHandler(ctx context.Context, token string) (encoreauth.UID, *UserData, error) {
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

return encoreauth.UID(userID), &UserData{
UserID: userID,
Role:   role,
Email:  email,
}, nil
}

func createToken(userID, email, role string) (string, error) {
token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
"user_id": userID,
"email":   email,
"role":    role,
"exp":     time.Now().Add(24 * time.Hour).Unix(),
})
return token.SignedString(jwtSecret)
}

var _ = uuid.UUID{}
var _ = sql.Open
