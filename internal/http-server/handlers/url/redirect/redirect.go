package redirect

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	resp "github.com/kxddry/url-shortener/internal/lib/api/response"
	"github.com/kxddry/url-shortener/internal/lib/logger/sl"
	"github.com/kxddry/url-shortener/internal/storage"
	"net/http"
)
import (
	"log/slog"
)

type URLGetSaver interface {
	URLGetter
	SaveURL(urlToSave, alias string) (int64, error)
}

type URLGetter interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlGetter URLGetter, redis URLGetSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.redirect.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Debug("alias is empty")
			w.WriteHeader(http.StatusNotAcceptable)
			render.JSON(w, r, resp.Error(resp.NotAcceptable, "alias is empty. Usage: POST to /url to create an alias or GET /{alias} to redirect"))
			return
		}
		resURL, err := redis.GetURL(alias) // check redis first
		if err != nil {
			resURL, err = urlGetter.GetURL(alias)
		}
		if errors.Is(err, storage.ErrAliasNotFound) {
			log.Debug("alias not found", slog.String("alias", alias))
			w.WriteHeader(http.StatusNotFound)
			render.JSON(w, r, resp.Error(resp.NotFound, "alias not found"))
			return
		}
		if err != nil {
			log.Error("failed to get URL", slog.String("alias", alias), sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(resp.InternalServerError, "failed to get URL"))
			return
		}
		log.Debug("alias found", slog.String("alias", alias), slog.String("url", resURL))
		_, err = redis.SaveURL(alias, resURL) // cache the URL in redis
		if err != nil {
			log.Error("failed to save URL in redis", slog.String("alias", alias), slog.String("url", resURL), sl.Err(err))
		}
		http.Redirect(w, r, resURL, http.StatusFound)
		log.Info("redirected", slog.String("alias", alias), slog.String("url", resURL))
		return
	}
}
