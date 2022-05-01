package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"urlshortener/handler"
	"urlshortener/storage"
	"urlshortener/urlshort"
)

func main() {
	filenameYaml := flag.String("yaml", "urls.yaml", "a yaml file containing shortened url and path")
	createInitData := flag.Bool("init", false, "specify true/false if initial db data should be created, you can specify from which file data should be loaded by json flag")
	filenameInitData := flag.String("json", "initdata.json", "a json file containing initial data for database")
	flag.Parse()

	boltDB := storage.NewBoltStorage("urlpath.db", 0666, nil)

	if err := boltDB.Connect(); err != nil {
		log.Fatal(err)
	}
	defer boltDB.Close()

	if *createInitData {
		if err := loadInitData(boltDB, *filenameInitData); err != nil {
			log.Fatal(err)
		}
	}

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
		"/tst":            "https://www.google.com",
	}
	mapHandler := handler.MapHandler(pathsToUrls, mux)

	yamlFileBytes, err := ioutil.ReadFile(*filenameYaml)
	if err != nil {
		log.Fatal(err)
	}

	yamlHandler, err := handler.YAMLHandler(yamlFileBytes, mapHandler)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", handler.DBHandler(boltDB, yamlHandler))
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}

func loadInitData(db *storage.BoltStorage, filenameInitData string) error {
	var urls []urlshort.PathUrl

	initDataBytes, err := ioutil.ReadFile(filenameInitData)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(initDataBytes, &urls); err != nil {
		return err
	}

	if err = db.CreateInitData(urls); err != nil {
		return err
	}

	return nil
}
