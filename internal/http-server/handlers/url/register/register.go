package register

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	cds "github.com/kxddry/sso-auth/output-error-codes"
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
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	resp.Response
	UserId int64 `json:"user_id,omitempty"`
}

type RegisterClient interface {
	Register(ctx context.Context, email, username, password string) (int64, error)
}

func New(ctx context.Context, log *slog.Logger, rc RegisterClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.register.New"

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
			render.JSON(w, r, resp.Error(resp.BadRequest, "failed to decode request. Correct usage: \"placeholder\": \"<email or username>\", \"password\": \"<password\""))
			return
		}

		log.Info("request body decoded", slog.String("email", req.Email))

		// better to validate without accessing SSO
		if v := validate(req, log, w, r); !v {
			return
		}

		uid, err := rc.Register(ctx, req.Email, req.Username, req.Password)
		if err != nil {
			if errors.Is(err, status.Error(codes.AlreadyExists, cds.UserAlreadyExists)) {
				log.Info("user already exists", sl.Err(err))
				w.WriteHeader(http.StatusBadRequest)
				render.JSON(w, r, resp.Error(resp.BadRequest, "user already exists"))
				return
			}

			log.Error("internal error", sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			render.JSON(w, r, resp.Error(resp.InternalServerError, "internal server error"))
			return
		}

		log.Info("user registered", slog.Int64("uid", uid))
		w.WriteHeader(http.StatusCreated)
		render.JSON(w, r, Response{
			Response: resp.OK(),
			UserId:   uid,
		})
		return
	}
}

func validate(req Request, log *slog.Logger, w http.ResponseWriter, r *http.Request) bool {
	if req.Email == "" {
		log.Debug("empty email")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "empty email"))
		return false
	}
	if req.Username == "" {
		log.Debug("empty username")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "empty username"))
		return false
	}
	if req.Password == "" {
		log.Debug("empty password")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "empty password"))
		return false
	}

	if !validator.ValidateEmail(req.Email) {
		log.Debug("invalid email", slog.String("email", req.Email))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "invalid email"))
		return false
	}

	if !validator.ValidateUsername(req.Username) {
		log.Debug("invalid username", slog.String("username", req.Email))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "invalid username"))
		return false
	}

	if !validator.ValidatePassword(req.Password) {
		log.Debug("invalid password", slog.String("password", req.Password))
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "invalid password"))
		return false
	}

	if len(req.Password) == 0 {
		log.Debug("empty password")
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, resp.Error(resp.BadRequest, "empty password"))
		return false
	}

	return true
}
