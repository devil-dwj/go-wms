package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	WmsDSN string
	Port   uint16
}

func NewConfig() *Config {
	return &Config{}
}

func MustLoad(path string) *Config {
	c := &Config{}
	c.LoadConfig(path)

	return c
}

func (c *Config) LoadConfig(file string) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	err = LoadConfigJson(b, c)
	if err != nil {
		panic(err)
	}
}

func LoadConfigJson(b []byte, v interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	return decoder.Decode(v)
}
