package storage

import (
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"golang.org/x/sync/errgroup"
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
		_, err = tx.CreateBucketIfNotExists([]byte(shortenedUrlBucket))

		return err
	})
}

func (boltStorage *BoltStorage) Close() error {
	return boltStorage.db.Close()
}

// InsertOne inserts PathUrl object into database.
func (boltStorage *BoltStorage) InsertOne(su *urlshort.PathUrl) (string, error) {
	fmt.Printf("Inserting %+v\n", su)

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

// FindOne looks for PathUrl object by given path. ShortenedUrlNotFoundErr is returned, if given path doesn't exist in database
func (boltStorage *BoltStorage) FindOne(path string) (*urlshort.PathUrl, error) {
	fmt.Printf("Looking for path: %v\n", path)
	var found urlshort.PathUrl

	if err := boltStorage.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(shortenedUrlBucket))
		if b == nil {
			return BucketDoesNotExistErr
		}

		urlBytes := b.Get([]byte(path))
		if urlBytes == nil {
			return urlshort.ShortenedUrlNotFoundErr
		}

		found = urlshort.PathUrl{
			Path: path,
			Url:  string(urlBytes),
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &found, nil
}

// CreateInitData creates initial data from given slice of PathUrl
func (boltStorage *BoltStorage) CreateInitData(urls []urlshort.PathUrl) error {
	g := new(errgroup.Group)

	for i := range urls {
		url := urls[i]

		g.Go(func() error {
			inserted, err := boltStorage.InsertOne(&url)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully inserted %v\n", inserted)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
