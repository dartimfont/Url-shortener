package update

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/domain"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type Request struct {
	NewURL string `json:"url" validate:"required,url"`
	Alias  string `json:"alias" validate:"required"`
}

type Response struct {
	resp.Response
	CountUpdated int64 `json:"countUpdated"`
}

//go:generate go run github.com/vektra/mockery/v2 --name=URLService
type URLService interface {
	UpdateURL(alias string, newURL string) (int64, error)
}

func New(log *slog.Logger, urlService URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.update.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("failed to decode request body"))
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

		alias := req.Alias
		newUrl := req.NewURL

		CountUpdated, err := urlService.UpdateURL(alias, newUrl)
		if errors.Is(err, domain.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, 404)
			render.JSON(w, r, resp.Error("url not found"))
			return
		}

		if err != nil {
			log.Error("failed to update url", "alias", alias)
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("failed to update url"))
			return
		}

		log.Info("url updated", slog.String("alias", alias), slog.Int64("count_updated", CountUpdated))

		responseOK(w, r, CountUpdated, 200)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, countUpdated int64, status int) {
	render.Status(r, status)
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		CountUpdated: countUpdated,
	})
}
