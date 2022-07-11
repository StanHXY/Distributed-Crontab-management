package master

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	ApiPort         int `json:"ApiPort"`
	ApiReadTimeout  int `json:"ApiReadTimeout"`
	ApiWriteTimeout int `json:"ApiWriteTimeout"`
}

var (
	G_config *Config
)

// load config
func InitConfig(filename string) (err error) {
	var (
		content []byte
		conf    Config
	)

	//1. read config file
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}

	//2. JSON Deserialize
	if err = json.Unmarshal(content, &conf); err != nil {
		return
	}

	// Assign
	G_config = &conf

	return
}
