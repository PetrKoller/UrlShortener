package urlshort

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

//
// Start of MapHandler testing
//

func TestMapHandler_Success(t *testing.T) {
	t.Parallel()

	pathsToUrls := map[string]string{
		"/tst": "https://www.google.com",
	}

	fallbackHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Fail(t, "Should've been redirected")
	})

	request, err := http.NewRequest(http.MethodGet, "/tst", nil)
	if err != nil {
		assert.Fail(t, "Creating GET request failed")
	}

	rr := httptest.NewRecorder()

	mapHandler := MapHandler(pathsToUrls, fallbackHandler)
	mapHandler.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusSeeOther, rr.Code)
}

func TestMapHandler_Fallback(t *testing.T) {
	t.Parallel()

	pathsToUrls := map[string]string{
		"/tst": "https://www.google.com",
	}

	fallbackHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("Fallback handler called"))
	})

	request, err := http.NewRequest(http.MethodGet, "/teest", nil)
	if err != nil {
		assert.Fail(t, "Creating GET request failed")
	}

	rr := httptest.NewRecorder()

	mapHandler := MapHandler(pathsToUrls, fallbackHandler)
	mapHandler.ServeHTTP(rr, request)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, []byte("Fallback handler called"), rr.Body.Bytes())
}

//
// End of MapHandler testing
//

//
// Start of YamlHandler testing
//

func TestYAMLHandler(t *testing.T) {
	t.Parallel()

	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution`

	fallbackHandlerTestFail := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Fail(t, "Should've been redirected")
	})

	fallbackHandlerCalled := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("Fallback handler called"))
	})

	tests := []struct {
		name            string
		path            string
		fallbackHandler http.Handler
		expectedCode    int
		fallbackCall    bool
	}{
		{
			name:            "Successfully redirected 1",
			path:            "/urlshort",
			fallbackHandler: fallbackHandlerTestFail,
			expectedCode:    http.StatusSeeOther,
			fallbackCall:    false,
		},
		{
			name:            "Successfully redirected 2",
			path:            "/urlshort-final",
			fallbackHandler: fallbackHandlerTestFail,
			expectedCode:    http.StatusSeeOther,
			fallbackCall:    false,
		},
		{
			name:            "Fallback called",
			path:            "/unknown-path",
			fallbackHandler: fallbackHandlerCalled,
			expectedCode:    http.StatusOK,
			fallbackCall:    true,
		},
	}

	for i := range tests {
		test := tests[i]

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request, err := http.NewRequest(http.MethodGet, test.path, nil)
			if err != nil {
				assert.Fail(t, "Creating GET request failed")
			}

			rr := httptest.NewRecorder()

			yamlHandler, err := YAMLHandler([]byte(yaml), test.fallbackHandler)
			assert.NoError(t, err)

			yamlHandler.ServeHTTP(rr, request)
			assert.Equal(t, test.expectedCode, rr.Code)

			if test.fallbackCall {
				assert.Equal(t, "Fallback handler called", rr.Body.String())
			}
		})
	}
}

func TestYAMLHandler_InvalidYaml(t *testing.T) {
	t.Parallel()

	xml := `<xml>
		<path>/urlshort</path>
<url>https://github.com/gophercises/urlshort</url>
</xml>
`

	_, err := YAMLHandler([]byte(xml), nil)

	assert.Error(t, err)
}

func TestYAMLHandler_AlreadyContainsPath(t *testing.T) {
	t.Parallel()

	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort
  url: https://github.com/gophercises/urlshort/tree/solution`

	_, err := YAMLHandler([]byte(yaml), nil)

	assert.ErrorIs(t, err, duplicatedPathErr)
}

//
// End of YamlHandler testing
//