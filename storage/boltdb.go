package storage

import (
	"errors"
	"github.com/boltdb/bolt"
	"os"
	"urlshortener/urlshort"
)

var BucketDoesNotExistErr = errors.New("bucket doesn't exist")

const shortenedUrlBucket string = "ShortenedUrlBucket"

type BoltStorage struct {
	db       *bolt.DB
	filepath string
	fileMode os.FileMode
	options  *bolt.Options
}

func NewBoltStorage(filepath string, fileMode os.FileMode, options *bolt.Options) *BoltStorage {
	return &BoltStorage{
		db:       &bolt.DB{},
		filepath: filepath,
		fileMode: fileMode,
		options:  options,
	}
}

func (boltStorage *BoltStorage) Connect() error {
	// TODO add logging
	var err error

	boltStorage.db, err = bolt.Open(boltStorage.filepath, boltStorage.fileMode, boltStorage.options)
	if err != nil {
		return err
	}

	return boltStorage.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(shortenedUrlBucket))

		return err
	})
}

func (boltStorage *BoltStorage) Close() error {
	return boltStorage.db.Close()
}

func (boltStorage *BoltStorage) InsertOne(su *urlshort.ShortenedUrl) (string, error) {
	if err := boltStorage.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(shortenedUrlBucket))
		if b == nil {
			return BucketDoesNotExistErr
		}

		return b.Put([]byte(su.Path), []byte(su.Url))
	}); err != nil {
		return "", err
	}

	return su.Path, nil
}

func (boltStorage *BoltStorage) FindOne(path string) (*urlshort.ShortenedUrl, error) {
	var found urlshort.ShortenedUrl

	if err := boltStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(shortenedUrlBucket))
		if b == nil {
			return BucketDoesNotExistErr
		}

		urlBytes := b.Get([]byte(path))
		if urlBytes == nil
	}); err != nil {

	}
}
