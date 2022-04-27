package main

import (
	"io"
	"io/ioutil"
)

func LoadYaml(reader io.Reader) ([]byte, error) {
	yamlFile, err := ioutil.ReadFile("conf.yaml")
	if err != nil {
		return nil, err
	}

	return yamlFile, nil
}
