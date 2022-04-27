package urlshort

type ShortenedUrlStorage interface {
	InsertOne(su *ShortenedUrl) (string, error)
	FindOne(path string) (*ShortenedUrl, error)
}

type ShortenedUrl struct {
	Path string `yaml:"path" json:"path"`
	Url  string `yaml:"url" json:"url"`
}
