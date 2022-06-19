package auth

import (
	"cmd/main/app.go/internal/config"
	"cmd/main/app.go/pkg/logging"
	"encoding/json"
	"github.com/cristalhq/jwt/v3"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"
)

const (
	authURL   = "/api/auth"
	signupURL = "/api/signup"
)

type user struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type refresh struct {
	RefreshToken string `json:"refresh_token"`
}

type Handler struct {
	Logger logging.Logger
	//UserService user_service.UserService
	//JWTHelper jwt.Helper
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodPost, URL, h.Auth)
	router.HandlerFunc(http.MethodPut, URL, h.Auth)
}

func (h *Handler) Auth(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	case http.MethodPost:
		var u user
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			h.Logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
	case http.MethodPut:
		var refreshTokenS refresh
		if err := json.NewDecoder(r.Body).Decode(&refreshTokenS); err != nil {
			h.Logger.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		userIdBytes, err := h.RTCache.Get([]byte(refreshTokenS.RefreshToken))
		h.Logger.Info("refresh token user_id: %s", userIdBytes)
		if err != nil {
			h.Logger.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		h.RTCache.Del([]byte(refreshTokenS.RefreshToken))
		// TODO отправка юзер в сервис создания пользователя
	}

	key := []byte(config.GetConfig().JWT.Secret)
	signer, err := jwt.NewSignerHS(jwt.HS256, key)
	if err != nil {
		w.WriteHeader(418)
		return
	}
	builder := jwt.NewBuilder(signer)

	// TODO заменить тестовые данные юзера
	claims := jwt2.UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "uuid",
			Audience:  []string{"users"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 1800)), //  30 мин на токен
		},
		Email: "email",
	}
	token, err := builder.Build(claims)
	if err != nil {
		h.Logger.Fatal(err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	jsonBytes, err := json.Marshal(map[string]string{
		"token":         token.String(),
		"refresh_token": refreshTokenUuid.String(),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
	w.WriteHeader(200)
}
