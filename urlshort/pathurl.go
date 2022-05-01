package urlshort

import "errors"

var ShortenedUrlNotFoundErr = errors.New("shortened url not found")

type PathUrlStorage interface {
	// InsertOne inserts PathUrl object into database.
	InsertOne(pu *PathUrl) (string, error)
	// FindOne looks for PathUrl object by given path. ShortenedUrlNotFoundErr is returned, if given path doesn't exist in database
	FindOne(path string) (*PathUrl, error)
}

type PathUrl struct {
	Path string `yaml:"path" json:"path"`
	Url  string `yaml:"url" json:"url"`
}
