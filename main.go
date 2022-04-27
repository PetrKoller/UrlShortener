package main

import (
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
	"log"
	"net/http"
	"urlshortener/urlshort"
)

func main() {
	filename := flag.String("yaml", "urls.yaml", "a yaml file containing shortened url and path")
	flag.Parse()

	db, err := bolt.Open("my.db", 0666, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("MyBucket"))
		if err != nil {
			return err
		}
		return b.Put([]byte("answer"), []byte("4445"))
	}); err != nil {
		log.Fatal(err)
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("MyBucket"))
		v := b.Get([]byte("answer"))
		fmt.Printf("The answer is: %s\n", v)
		return nil
	})

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
		"/tst":            "https://www.google.com",
	}
	mapHandler := urlshort.MapHandler(pathsToUrls, mux)

	yamlFileBytes, err := ioutil.ReadFile(*filename)
	if err != nil {
		log.Fatal(err)
	}

	yamlHandler, err := urlshort.YAMLHandler(yamlFileBytes, mapHandler)
	if err != nil {
		panic(err)
	}
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", yamlHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
