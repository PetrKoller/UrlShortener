package handler

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"urlshortener/storage"
	"urlshortener/urlshort"
)

//
// Start of DBHandler testing
// integration tests
//

var dbFileLocation = "../testingData/test.db"

func tearDown(t *testing.T) {
	if err := os.Remove(dbFileLocation); err != nil {
		t.Fatal(err)
	}
}

func TestDBHandler(t *testing.T) {
	db := storage.NewBoltStorage(dbFileLocation, 0666, nil)

	if err := db.Connect(); err != nil {
		assert.Fail(t, "Couldn't establish connection with DB")
	}
	defer db.Close()
	defer tearDown(t)

	pathUrl1 := urlshort.PathUrl{
		Path: "/test1",
		Url:  "https://www.facebook.com",
	}

	pathUrl2 := urlshort.PathUrl{
		Path: "/test2",
		Url:  "https://www.youtube.com",
	}
	if err := db.CreateInitData([]urlshort.PathUrl{pathUrl1, pathUrl2}); err != nil {
		assert.Fail(t, "Creating initial data failed")
	}

	tests := []struct {
		name           string
		pathUrl        urlshort.PathUrl
		shouldRedirect bool
		header         int
		responseBody   string
	}{
		{
			name:           "successfully_redirected_1",
			pathUrl:        pathUrl1,
			shouldRedirect: true,
			header:         http.StatusSeeOther,
			responseBody:   fmt.Sprintf("<a href=\"%v\">See Other</a>.\n\n", pathUrl1.Url),
		},
		{
			name:           "successfully_redirected_2",
			pathUrl:        pathUrl2,
			shouldRedirect: true,
			header:         http.StatusSeeOther,
			responseBody:   fmt.Sprintf("<a href=\"%v\">See Other</a>.\n\n", pathUrl2.Url),
		},
		{
			name: "nonexisting_path",
			pathUrl: urlshort.PathUrl{
				Path: "/randomPath",
				Url:  "www.rangomurl.org",
			},
			shouldRedirect: false,
			header:         http.StatusOK,
			responseBody:   "Fallback handler called",
		},
	}

	fallbackHandler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Write([]byte("Fallback handler called"))
		response.WriteHeader(http.StatusOK)
	})

	dbHandler := DBHandler(db, fallbackHandler)

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, test.pathUrl.Path, nil)
			if err != nil {
				assert.Fail(t, "Creating GET request failed")
			}

			rr := httptest.NewRecorder()
			dbHandler.ServeHTTP(rr, request)

			if test.shouldRedirect {
				assert.Equal(t, http.StatusSeeOther, rr.Code)
				assert.Equal(t, test.responseBody, rr.Body.String())
			} else {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Equal(t, "Fallback handler called", rr.Body.String())
			}
		})
	}
}

func TestDBHandler_DBError(t *testing.T) {
	db := storage.NewBoltStorage(dbFileLocation, 0666, nil)
	// connection not established to easily produce error, (mocking up db storage is unnecessary for this case, in my opinion)

	request, err := http.NewRequest(http.MethodGet, "/randomTest", nil)
	if err != nil {
		assert.Fail(t, "Creating GET request failed")
	}

	rr := httptest.NewRecorder()
	DBHandler(db, nil).ServeHTTP(rr, request)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "Database error\n", rr.Body.String())
}

//
// End of DBHandler testing
//

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
// Start of YAMLHandler testing
//

func TestYAMLHandler_Redirect(t *testing.T) {
	t.Parallel()

	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution`

	testHandlerRedirect(t, []byte(yaml), YAMLHandler)
}

func TestYAMLHandler_InvalidYaml(t *testing.T) {
	t.Parallel()

	xml := `<xml>
		<path>/urlshort</path>
<url>https://github.com/gophercises/urlshort</url>
</xml>
`

	testInvalidData(t, []byte(xml), YAMLHandler)
}

func TestYAMLHandler_AlreadyContainsPath(t *testing.T) {
	t.Parallel()

	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort
  url: https://github.com/gophercises/urlshort/tree/solution`

	testDuplicatedPath(t, []byte(yaml), YAMLHandler)
}

//
// End of YAMLHandler testing
//

//
// Start of JSONHandler testing
//

func TestJSONHandler_Redirect(t *testing.T) {
	t.Parallel()

	json := `
[
	{
		"path": "/urlshort",
		"url": "https://github.com/gophercises/urlshort"
	},
	{
		"path": "/urlshort-final",
		"url": "https://github.com/gophercises/urlshort/tree/solution"
	}
]`

	testHandlerRedirect(t, []byte(json), JSONHandler)
}

func TestJSONHandler_InvalidJSON(t *testing.T) {
	t.Parallel()

	xml := `<xml>
		<path>/urlshort</path>
<url>https://github.com/gophercises/urlshort</url>
</xml>
`

	testInvalidData(t, []byte(xml), JSONHandler)
}

func TestJSONHandler_AlreadyContainsPath(t *testing.T) {
	t.Parallel()

	json := `
[
	{
		"path": "/urlshort",
		"url": "https://github.com/gophercises/urlshort"
	},
	{
		"path": "/urlshort",
		"url": "https://github.com/gophercises/urlshort/tree/solution"
	}
]`

	testDuplicatedPath(t, []byte(json), JSONHandler)
}

//
// End of JSONHandler testing
//

func testHandlerRedirect(t *testing.T, dataBytes []byte, handler func(dataBytes []byte, fallback http.Handler) (http.HandlerFunc, error)) {
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

			handlerToTest, err := handler(dataBytes, test.fallbackHandler)
			assert.NoError(t, err)

			handlerToTest.ServeHTTP(rr, request)
			assert.Equal(t, test.expectedCode, rr.Code)

			if test.fallbackCall {
				assert.Equal(t, "Fallback handler called", rr.Body.String())
			}
		})
	}
}

func testInvalidData(t *testing.T, invalidDataBytes []byte, handler func(ymlBytes []byte, fallback http.Handler) (http.HandlerFunc, error)) {
	_, err := handler(invalidDataBytes, nil)

	assert.Error(t, err)
}

func testDuplicatedPath(t *testing.T, dataBytes []byte, handler func(ymlBytes []byte, fallback http.Handler) (http.HandlerFunc, error)) {
	_, err := handler(dataBytes, nil)

	assert.ErrorIs(t, err, DuplicatedPathErr)
}
