package redirect_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"url-shortener/internal/handlers/url/redirect"
	"url-shortener/internal/handlers/url/redirect/mocks"
	"url-shortener/internal/infra/api"
	"url-shortener/internal/infra/logger/handlers/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	resp "url-shortener/internal/infra/api/response"
)

func TestRedirect_Success(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test_alias",
			url:   "https://www.google.com/",
		},
		{
			name:  "Success",
			alias: "google",
			url:   "https://www.google.com/",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			urlServiceMock := mocks.NewURLService(t)

			if tc.respError == "" || tc.mockError != nil {
				urlServiceMock.On("GetURL", tc.alias).
					Return(tc.url, nil).
					Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlServiceMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			assert.Equal(t, tc.url, redirectedToURL)
		})
	}
}

func TestRedirect_BadRequest(t *testing.T) {
	urlServiceMock := mocks.NewURLService(t)

	handler := redirect.New(slogdiscard.NewDiscardLogger(), urlServiceMock)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusBadRequest)

	var resp resp.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "invalid request", resp.Error)

	urlServiceMock.AssertNotCalled(t, "GetURL", mock.Anything)
}
