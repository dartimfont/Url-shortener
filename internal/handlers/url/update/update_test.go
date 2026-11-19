package update_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/handlers/url/update"
	"url-shortener/internal/handlers/url/update/mocks"
	"url-shortener/internal/infra/logger/handlers/slogdiscard"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdate_Success(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		newURL       string
		countUpdated int64
	}{
		{
			name:         "Success with alias",
			newURL:       "https://google.com",
			alias:        "test_alias",
			countUpdated: 1,
		},
		{
			name:         "Success with alias 2",
			newURL:       "https://vk.ru",
			alias:        "test_alias",
			countUpdated: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlServiceMock := mocks.NewURLService(t)
			urlServiceMock.On("UpdateURL", tc.alias, tc.newURL).
				Return(int64(1), nil).
				Once()

			handler := update.New(slogdiscard.NewDiscardLogger(), urlServiceMock)

			body, err := json.Marshal(update.Request{NewURL: tc.newURL, Alias: tc.alias})
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/url", bytes.NewReader(body))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusOK)

			var resp update.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Empty(t, resp.Error)
			require.Equal(t, tc.countUpdated, resp.CountUpdated)
		})
	}
}

func TestUpdate_BadRequest(t *testing.T) {
	cases := []struct {
		name      string
		alias     string
		newURL    string
		rawBody   string
		respError string
	}{
		{
			name:      "Empty URL",
			alias:     "some_alias",
			newURL:    "",
			respError: "field NewURL is required field",
		},
		{
			name:      "Invalid URL",
			alias:     "some_alias",
			newURL:    "some invalid URL",
			respError: "field NewURL is not a valid URL",
		},
		{
			name:      "Invalid JSON",
			rawBody:   `{"url":`,
			respError: "failed to decode request body",
		},
		{
			name:      "Something wrong",
			rawBody:   `1`,
			respError: "failed to decode request body",
		},
		{
			name:      "Empty alias",
			alias:     "",
			newURL:    "https://google.com",
			respError: "field Alias is required field",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlServiceMock := mocks.NewURLService(t)
			handler := update.New(slogdiscard.NewDiscardLogger(), urlServiceMock)

			var body []byte
			if tc.rawBody != "" {
				body = []byte(tc.rawBody)
			} else {
				var err error
				body, err = json.Marshal(update.Request{NewURL: tc.newURL, Alias: tc.alias})
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, "/url", bytes.NewReader(body))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			require.Equal(t, rr.Code, http.StatusBadRequest)

			var resp update.Response
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
			require.Equal(t, resp.Error, tc.respError)

			urlServiceMock.AssertNotCalled(t, "UpdateURL", mock.Anything, mock.Anything)
		})
	}
}
