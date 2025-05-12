package save

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/kxddry/url-shortener/internal/config"
	resp "github.com/kxddry/url-shortener/internal/lib/api/response"
	"github.com/kxddry/url-shortener/internal/lib/genalias"
	"github.com/kxddry/url-shortener/internal/lib/jwt"
	"github.com/kxddry/url-shortener/internal/lib/logger/sl"
	"github.com/kxddry/url-shortener/internal/storage"
	"io"
	"log/slog"
	"net/http"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave, alias string, creator int64) (int64, error)
}

type URLGetter interface {
	GetURL(alias string) (string, error)
}

type URLSaveGetter interface {
	URLSaver
	URLGetter
}

const aliasLength = 6

func New(log *slog.Logger, urlSaver URLSaveGetter, redis URLSaver, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Error("request body is empty", sl.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error(resp.BadRequest, "request body is empty"))
				return
			}
			log.Error("failed to decode request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.BadRequest, "failed to decode request"))
			return
		}

		uid, err := jwt.UIDfromHeader(r, cfg.App.Secret)

		if err != nil {
			if errors.Is(err, jwt.ErrNoHeader) {
				log.Info("not logged in")
				w.WriteHeader(http.StatusUnauthorized)
				render.JSON(w, r, resp.Error(resp.Unauthorized, "You're not logged in. Go to /login."))
				return
			}

			log.Error("invalid token", slog.Any("jwt", r.Header.Get("Authorization")))
			w.WriteHeader(http.StatusUnauthorized)
			render.JSON(w, r, resp.Error(resp.Unauthorized, "invalid token"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		alias := req.Alias
		if alias == "" {
			var err error
			alias, err = genalias.GenerateAlias(aliasLength, urlSaver)
			if err != nil {
				log.Error("failed to generate alias", sl.Err(err))
				w.WriteHeader(http.StatusInternalServerError)
				render.JSON(w, r, resp.Error(resp.InternalServerError, "failed to generate alias"))
				return
			}
			log.Info("generated alias", slog.String("alias", alias))
		}
		if alias == "url" {
			log.Error("alias is reserved", slog.String("alias", alias))
			w.WriteHeader(http.StatusNotAcceptable)
			render.JSON(w, r, resp.Error(resp.NotAcceptable, "alias is reserved"))
			return
		}

		// save to redis first
		_, err = redis.SaveURL(req.URL, alias, uid)
		if err != nil {
			log.Error(fmt.Sprintf("failed to save to redis %s: %w", op, err))
		}

		id, err := urlSaver.SaveURL(req.URL, alias, uid)
		if errors.Is(err, storage.ErrAliasExists) {
			log.Error("alias already exists", sl.Err(err))
			w.WriteHeader(http.StatusNotAcceptable)
			render.JSON(w, r, resp.Error(resp.NotAcceptable, "alias already exists"))
			return
		}
		if err != nil {
			log.Error("failed to save url", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(resp.InternalServerError, "failed to save url"))
			return
		}
		log.Info("url saved", slog.Int64("id", id), slog.String("alias", alias))
		responseOK(w, r, alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
