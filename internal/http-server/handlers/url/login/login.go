package login

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	cds "github.com/kxddry/sso-auth/output-error-codes"
	"github.com/kxddry/url-shortener/internal/config"
	resp "github.com/kxddry/url-shortener/internal/lib/api/response"
	"github.com/kxddry/url-shortener/internal/lib/logger/sl"
	"github.com/kxddry/url-shortener/internal/lib/validator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
	"net/http"
)

type Request struct {
	Placeholder string `json:"placeholder" validate:"required"` // username or email
	Password    string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
}

type LoginClient interface {
	Login(ctx context.Context, placeholder, pass string, appId int64) (string, error)
}

func New(ctx context.Context, log *slog.Logger, cfg *config.Config, lc LoginClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.login.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())))

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			if errors.Is(err, io.EOF) {
				log.Debug("request body is empty", sl.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error(resp.BadRequest, "request body is empty"))
				return
			}
			log.Error("failed to decode request", sl.Err(err))
			w.WriteHeader(http.StatusBadRequest)
			render.JSON(w, r, resp.Error(resp.BadRequest, `failed to decode request. Correct usage: "placeholder": "(email or username)", "password": "(password)"`))
			return
		}

		log.Info("request body decoded", slog.String("placeholder", req.Placeholder))

		if v := validate(req, log, w, r); !v {
			return
		}

		token, err := lc.Login(ctx, req.Placeholder, req.Password, cfg.App.ID)
		if err != nil {
			if errors.Is(err, status.Error(codes.InvalidArgument, cds.InvalidCredentials)) {
				log.Debug("invalid credentials", sl.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error(resp.BadRequest, "invalid credentials. Try registering at /register."))
				return
			}

			log.Error(cds.InternalError, sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(resp.InternalServerError, "internal server error"))
			return
		}

		// user is logged in without any errors
		response := &Response{
			Response:    resp.OK(),
			AccessToken: token,
			TokenType:   "Bearer",
			ExpiresIn:   int(cfg.TokenTTL.Seconds()),
		}

		w.WriteHeader(http.StatusOK)
		render.JSON(w, r, response)
	}
}

func validate(req Request, log *slog.Logger, w http.ResponseWriter, r *http.Request) bool {
	if req.Placeholder == "" {
		log.Debug("empty placeholder")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "empty placeholder"))
		return false
	}
	if req.Password == "" {
		log.Debug("empty password")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "empty password"))
		return false
	}
	if !validateRequest(req) {
		log.Debug("bad request", slog.String("placeholder", req.Placeholder))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "invalid placeholder"))
		return false
	}
	return true
}

const (
	Fail     = validator.Fail
	Username = validator.Username
	Email    = validator.Email
)

func validateRequest(req Request) bool {
	return validator.ValidatePlaceholder(req.Placeholder) >= 0
}
