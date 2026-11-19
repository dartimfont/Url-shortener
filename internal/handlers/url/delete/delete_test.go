package delete_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/handlers/url/delete/mocks"
	"url-shortener/internal/infra/logger/handlers/slogdiscard"

	del "url-shortener/internal/handlers/url/delete"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDelete_Success(t *testing.T) {
	cases := []struct {
		name         string
		alias        string
		countDeleted int64
	}{
		{
			name:         "Success with alias",
			alias:        "test_alias_1",
			countDeleted: 1,
		},
		{
			name:         "Success with alias 2",
			alias:        "test_alias_2",
			countDeleted: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlServiceMock := mocks.NewURLService(t)
			urlServiceMock.On("DeleteURL", tc.alias).
				Return(tc.countDeleted, nil).
				Once()

			r := chi.NewRouter()
			r.Delete("/{alias}", del.New(slogdiscard.NewDiscardLogger(), urlServiceMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/"+tc.alias, nil)
			require.NoError(t, err)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, resp.StatusCode, http.StatusOK)

			var delResponse del.Response
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&delResponse))
			require.Empty(t, delResponse.Error)
			require.Equal(t, tc.countDeleted, delResponse.CountDeleted)
		})
	}
}

func TestDelete_BadRequest(t *testing.T) {
	urlServiceMock := mocks.NewURLService(t)

	handler := del.New(slogdiscard.NewDiscardLogger(), urlServiceMock)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	require.Equal(t, rr.Code, http.StatusBadRequest)

	var resp del.Response
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	require.Equal(t, "invalid request", resp.Error)

	urlServiceMock.AssertNotCalled(t, "DeleteURL", mock.Anything)
}
