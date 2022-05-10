package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

func MustLoad(path string, v interface{}) {
	if err := LoadConfig(path, v); err != nil {
		panic(err)
	}
}

func LoadConfig(file string, v interface{}) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = LoadConfigJson(b, v)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfigJson(b []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	return decoder.Decode(v)
}
