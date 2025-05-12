package homepage

import (
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/kxddry/url-shortener/internal/config"
	resp "github.com/kxddry/url-shortener/internal/lib/api/response"
	"github.com/kxddry/url-shortener/internal/lib/jwt"
	"github.com/kxddry/url-shortener/internal/lib/logger/sl"
	"log/slog"
	"net/http"
)

func Url(log *slog.Logger, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.homepage.Url"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Debug("visited website")

		_, err := jwt.UIDfromHeader(r, cfg.App.Secret)

		if err != nil {
			log.Debug("problem with jwt token", sl.Err(err))
			switch {
			case errors.Is(err, jwt.ErrNoHeader):
				http.Redirect(w, r, "/login", http.StatusSeeOther)
			case errors.Is(err, jwt.ErrInvalidHeader):
				http.Error(w, jwt.ErrInvalidHeader.Error(), http.StatusBadRequest)
			default:
				http.Error(w, err.Error(), http.StatusNotAcceptable)
			}

			return
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, resp.Info("Welcome to the URL shortener! Usage: POST to /url to create an alias; GET /{alias} to get redirected; DELETE /{alias} to delete the alias. Register and log in at /register and /login."))
	}
}
