package urlshort

import (
	"encoding/json"
	"errors"
	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v2"
	"net/http"
)

var DuplicatedPathErr = errors.New("duplicated path, path has already got assigned url")

func DBHandler(db *bolt.DB, fallback http.Handler) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {

	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//     - path: /some-path
//       url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data or DuplicatedPathErr .
//
func YAMLHandler(ymlBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {
	yamlParsed, err := parseYaml(ymlBytes)
	if err != nil {
		return nil, err
	}

	pathMap, err := buildMap(yamlParsed)
	if err != nil {
		return nil, err
	}

	return MapHandler(pathMap, fallback), nil
}

// JSONHandler will parse the provided JSON and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the JSON, then the
// fallback http.Handler will be called instead.
//
// JSON is expected to be in the format:
//
//     [
//		  	{
//				"path": "/some-path",
//				"url": "https://www.some-url.com/demo",
//			}
//     ]
//
// The only errors that can be returned all related to having
// invalid JSON data.
//
func JSONHandler(jsonBytes []byte, fallback http.Handler) (http.HandlerFunc, error) {
	jsonParsed, err := parseJson(jsonBytes)
	if err != nil {
		return nil, err
	}

	pathMap, err := buildMap(jsonParsed)
	if err != nil {
		return nil, err
	}

	return MapHandler(pathMap, fallback), nil
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(response http.ResponseWriter, request *http.Request) {
		val, ok := pathsToUrls[request.URL.Path]
		if !ok {
			fallback.ServeHTTP(response, request)
			return
		}
		http.Redirect(response, request, val, http.StatusSeeOther)
	}
}

func parseYaml(ymlBytes []byte) ([]ShortenedUrl, error) {
	var result []ShortenedUrl

	err := yaml.Unmarshal(ymlBytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func parseJson(jsonBytes []byte) ([]ShortenedUrl, error) {
	var result []ShortenedUrl

	err := json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func buildMap(shortenedUrls []ShortenedUrl) (map[string]string, error) {
	urlMap := make(map[string]string, len(shortenedUrls))

	for _, shortened := range shortenedUrls {
		if _, ok := urlMap[shortened.Path]; ok {
			return nil, DuplicatedPathErr
		}

		urlMap[shortened.Path] = shortened.Url
	}

	return urlMap, nil
}
