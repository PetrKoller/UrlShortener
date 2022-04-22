package urlshort

import (
	"errors"
	"gopkg.in/yaml.v2"
	"net/http"
)

var duplicatedPathErr = errors.New("duplicated path, path has already got assigned url")

type shortenedUrl struct {
	Path string `yaml:"path"`
	Url  string `yaml:"url"`
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
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	yamlParsed, err := parseYaml(yml)
	if err != nil {
		return nil, err
	}

	pathMap, err := buildMap(yamlParsed)
	if err != nil {
		return nil, err
	}

	return MapHandler(pathMap, fallback), nil
}

func parseYaml(yml []byte) ([]shortenedUrl, error) {
	var result []shortenedUrl

	err := yaml.Unmarshal(yml, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func buildMap(shortenedUrls []shortenedUrl) (map[string]string, error) {
	urlMap := make(map[string]string, len(shortenedUrls))

	for _, shortened := range shortenedUrls {
		if _, ok := urlMap[shortened.Path]; ok {
			return nil, duplicatedPathErr
		}

		urlMap[shortened.Path] = shortened.Url
	}

	return urlMap, nil
}
