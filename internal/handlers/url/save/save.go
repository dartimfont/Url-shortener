package save

import (
	"errors"
	"log/slog"
	"net/http"

	"url-shortener/internal/domain"
	resp "url-shortener/internal/infra/api/response"
	"url-shortener/internal/infra/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	resp.Response
	Alias string `json:"alias,omitempty"`
}

//go:generate go run github.com/vektra/mockery/v2 --name=URLService
type URLService interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlService URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.save.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("failed to decode request"))
			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Error("invalid request", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, resp.ValidationError(validateErr))
			return
		}

		id, err := urlService.SaveURL(req.URL, req.Alias)
		if errors.Is(err, domain.ErrURLExists) {
			log.Info("url already exists", slog.String("url", req.URL))
			render.Status(r, 409)
			render.JSON(w, r, resp.Error("url already exists"))
			return
		}
		if err != nil {
			log.Error("failed to add url", sl.Err(err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("failed to add url"))
			return
		}

		log.Info("url added", slog.Int64("id", id))

		responseOK(w, r, req.Alias, 201)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string, status int) {
	render.Status(r, status)
	render.JSON(w, r, Response{
		Response: resp.OK(),
		Alias:    alias,
	})
}
