package redirect

import (
	"errors"
	"log/slog"
	"net/http"

	"url-shortener/internal/domain"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/sl"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

//go:generate go run github.com/vektra/mockery/v2 --name=URLService
type URLService interface {
	GetURL(alias string) (string, error)
}

func New(log *slog.Logger, urlService URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.redirect.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Info("alias is empty")
			render.Status(r, 400)
			render.JSON(w, r, resp.Error("invalid request"))
			return
		}

		resURL, err := urlService.GetURL(alias)
		if errors.Is(err, domain.ErrURLNotFound) {
			log.Info("url not found", "alias", alias)
			render.Status(r, 404)
			render.JSON(w, r, resp.Error("not found"))
			return
		}

		if err != nil {
			log.Error("failed to get url", sl.Err(err))
			render.Status(r, 500)
			render.JSON(w, r, resp.Error("internal error"))
			return
		}

		log.Info("got url", slog.String("url", resURL))

		http.Redirect(w, r, resURL, http.StatusFound)
	}
}
