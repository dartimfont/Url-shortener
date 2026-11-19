package save_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/domain"
	"url-shortener/internal/handlers/url/save"
	"url-shortener/internal/handlers/url/save/mocks"
	"url-shortener/internal/infra/logger/handlers/slogdiscard"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSave_Success(t *testing.T) {
	cases := []struct {
		name  string
		alias string
		url   string
	}{
		{
			name:  "Success with alias",
			alias: "google_alias",
			url:   "https://google.com",
		},
		{
			name:  "Success with alias 2",
			alias: "google_alias_1",
			url:   "https://google.com",
		},
		{
			name:  "Success with empty alias",
			alias: "",
			url:   "https://google.com",
		},
		{
			name:  "Success deepseek with empty alias",
			alias: "",
			url:   "https://chat.deepseek.com/",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlServiceMock := mocks.NewURLService(t)
			urlServiceMock.On("SaveURL", tc.url, tc.alias).
				Return(int64(1), nil).
				Once()

			handler := save.New(slogdiscard.NewDiscardLogger(), urlServiceMock)

			body, err := json.Marshal(save.Request{URL: tc.url, Alias: tc.alias})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader(body))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusCreated)

			var resp save.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Empty(t, resp.Error)
			require.Equal(t, tc.alias, resp.Alias)
		})
	}
}

func TestSave_BadRequest(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		rawBody   string
		respError string
	}{
		{
			name:      "Empty URL",
			alias:     "some_alias",
			url:       "",
			respError: "field URL is required field",
		},
		{
			name:      "Invalid URL",
			alias:     "some_alias",
			url:       "some invalid URL",
			respError: "field URL is not a valid URL",
		},
		{
			name:      "Invalid JSON",
			rawBody:   `{"url":`,
			respError: "failed to decode request",
		},
		{
			name:      "Something wrong",
			rawBody:   `1`,
			respError: "failed to decode request",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlServiceMock := mocks.NewURLService(t)
			handler := save.New(slogdiscard.NewDiscardLogger(), urlServiceMock)

			var body []byte
			if tc.rawBody != "" {
				body = []byte(tc.rawBody)
			} else {
				var err error
				body, err = json.Marshal(save.Request{URL: tc.url, Alias: tc.alias})
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader(body))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusBadRequest)

			var resp save.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Equal(t, resp.Error, tc.respError)

			urlServiceMock.AssertNotCalled(t, "SaveURL", mock.Anything, mock.Anything)
		})
	}
}

func TestSave_Conflict(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:      "Alias already exists",
			alias:     "test_alias",
			url:       "https://google.com",
			respError: "url already exists",
		},
		{
			name:      "Alias already exists for another url",
			alias:     "test_alias",
			url:       "https://hh.ru",
			respError: "url already exists",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			urlServiceMock := mocks.NewURLService(t)
			urlServiceMock.On("SaveURL", tc.url, tc.alias).
				Return(int64(0), domain.ErrURLExists).
				Once()

			handler := save.New(slogdiscard.NewDiscardLogger(), urlServiceMock)

			body, err := json.Marshal(save.Request{URL: tc.url, Alias: tc.alias})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/url", bytes.NewReader(body))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusConflict)

			var resp save.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Equal(t, resp.Error, tc.respError)
			require.Empty(t, resp.Alias)
		})
	}
}
