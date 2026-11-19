package delete

import (
	"errors"
	"log/slog"
	"net/http"
	"url-shortener/internal/domain"
	resp "url-shortener/internal/lib/api/response"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	resp.Response
	CountDeleted int64 `json:"countDeleted"`
}

//go:generate go run github.com/vektra/mockery/v2 --name=URLService
type URLService interface {
	DeleteURL(alias string) (int64, error)
}

func New(log *slog.Logger, urlService URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.delete.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("empty alias")
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		countDeleted, err := urlService.DeleteURL(alias)
		if errors.Is(err, domain.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, 404)
			render.JSON(w, r, resp.Error("url not fount"))
			return
		}
		if err != nil {
			log.Error("failed to delete url", "alias", alias)
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("deleted url", slog.String("alias", alias), slog.Int64("count_deleted", countDeleted))

		responseOK(w, r, countDeleted, 200)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, countDeleted int64, status int) {
	render.Status(r, status)
	render.JSON(w, r, Response{
		Response:     resp.OK(),
		CountDeleted: countDeleted,
	})
}
