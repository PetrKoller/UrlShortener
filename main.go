package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"urlshortener/storage"
	"urlshortener/urlshort"
)

func main() {
	filename := flag.String("yaml", "urls.yaml", "a yaml file containing shortened url and path")
	flag.Parse()

	boltDB := storage.NewBoltStorage("test.db", 0666, nil)

	if err := boltDB.Connect(); err != nil {
		log.Fatal(err)
	}
	defer boltDB.Close()

	_, _ = boltDB.InsertOne(&urlshort.ShortenedUrl{Path: "test", Url: "https://www.google.com"})

	su, err := boltDB.FindOne("tedsst")
	if err == storage.ShortenedUrlNotFound {
		fmt.Printf("Not found")
	} else {
		fmt.Printf("%+v", su)
	}

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
