package metric

import (
	"cmd/main/app.go/pkg/logging"
	"cmd/main/app.go/pkg/middleware/jwt"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

const (
	URL = "/api/heartbeat"
)

type Handler struct {
	Logger logging.Logger
}

func (h *Handler) Register(router *httprouter.Router) {
	router.HandlerFunc(http.MethodGet, URL, jwt.JWTMiddleware(h.Heartbeat))
}

func (h *Handler) Heartbeat(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(204)
}
