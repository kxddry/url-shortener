package delete

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/kxddry/url-shortener/internal/config"
	resp "github.com/kxddry/url-shortener/internal/lib/api/response"
	"github.com/kxddry/url-shortener/internal/lib/jwt"
	"github.com/kxddry/url-shortener/internal/lib/logger/sl"
	"github.com/kxddry/url-shortener/internal/storage"
	"log/slog"
	"net/http"
)

type URLDeleter interface {
	DeleteURL(alias string) error
}

type CreatorFinder interface {
	Creator(alias string) (int64, error)
}

type Storage interface {
	URLDeleter
	CreatorFinder
}

type AdminChecker interface {
	IsAdmin(ctx context.Context, userID int64) (bool, error)
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config, store Storage, redis URLDeleter, sso AdminChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Debug("alias is empty")
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.BadRequest, "alias is empty"))
			return
		}

		uid, err := jwt.UIDfromHeader(r, cfg.App.Secret)
		if err != nil {
			if errors.Is(err, jwt.ErrNoHeader) {
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, resp.Error(resp.Unauthorized, "go to /login or /register"))
				return
			}

			if errors.Is(err, jwt.ErrInvalidToken) {
				log.Info("invalid token")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error(resp.BadRequest, "invalid token"))
				return
			}

			if errors.Is(err, jwt.ErrInvalidHeader) {
				log.Info("invalid header")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error(resp.BadRequest, "invalid header"))
				return
			}

			log.Error("internal error!", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(resp.InternalServerError, "internal server error"))
			return
		}

		creator, err := store.Creator(alias)
		if err != nil {
			if errors.Is(err, storage.ErrAliasNotFound) {
				log.Info("alias not found")
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error(resp.BadRequest, "alias not found"))
				return
			}

			log.Error("internal error!", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(resp.InternalServerError, "internal server error"))
			return
		}

		if creator == uid {
			delete(log, store, alias, redis, w, r)
			return
		}

		isAdmin, err := sso.IsAdmin(ctx, uid)
		if err != nil {
			log.Error("internal error!", sl.Err(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if isAdmin {
			delete(log, store, alias, redis, w, r)
			return
		}

		log.Info("user tried to delete alias", slog.Int64("uid", uid))
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		return
	}
}

func delete(log *slog.Logger, store Storage, alias string, redis URLDeleter, w http.ResponseWriter, r *http.Request) {
	err := store.DeleteURL(alias)
	if err != nil {
		log.Error("internal error!", sl.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = redis.DeleteURL(alias)
	if err != nil {
		log.Error("internal error!", sl.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	render.JSON(w, r, resp.OK())
	return
}
