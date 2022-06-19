package jwt

import (
	"cmd/main/app.go/internal/app_context"
	"cmd/main/app.go/pkg/logging"
	"context"
	"encoding/json"
	"github.com/cristalhq/jwt/v3"
	"net/http"
	"strings"
	"time"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

func JWTMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := logging.GetLogger()
		authHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")
		if len(authHeader) != 2 {
			logger.Error("Неправильно сформированный токен")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Неправильно сформированный токен"))
			return
		}
		logger.Debug("Создаём JWT...")
		jwtToken := authHeader[1]
		key := []byte(app_context.GetInstance().Config.JWT.Secret)
		verifier, err := jwt.NewVerifierHS(jwt.HS256, key)
		if err != nil {
			unauthorized(w, err)
			return
		}
		logger.Debug("Парсим и проверяем JWT токен")
		token, err := jwt.ParseAndVerifyString(jwtToken, verifier)
		if err != nil {
			unauthorized(w, err)
			return
		}

		logger.Debug("Получаем JWT пользователя")
		var uc UserClaims
		err = json.Unmarshal(token.RawClaims(), &uc)
		if err != nil {
			unauthorized(w, err)
			return
		}
		if valid := uc.IsValidAt(time.Now()); !valid {
			logger.Error("Токен протух")
			unauthorized(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", uc.ID)
		h(w, r.WithContext(ctx))
	}
}

func unauthorized(w http.ResponseWriter, err error) {
	logging.GetLogger().Error(err)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("unauthorized"))
}
