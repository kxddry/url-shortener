package homepage

import (
	"github.com/go-chi/render"
	"github.com/kxddry/url-shortener/internal/config"
	resp "github.com/kxddry/url-shortener/internal/lib/api/response"
	"github.com/kxddry/url-shortener/internal/lib/jwt"
	"net/http"
)

func Register(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := jwt.UIDfromHeader(r, cfg.App.Secret)
		if err == nil {
			render.JSON(w, r, resp.Info("redirecting to /, you're logged in"))
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, resp.Info("POST a JSON to register: email, username, password"))
	}
}
