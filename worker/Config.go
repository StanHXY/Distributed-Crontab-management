package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	EtcdEndpoints         []string `json:"etcdEndpoints"`
	EtcdDialTimeout       int      `json:"etcdDialTimeout"`
	MongodbUri            string   `json:"mongodbUri"`
	MongodbConnectTimeout int      `json:"mongodbConnectTimeout"`
	JobLogBatchSize       int      `json:"jobLogBatchSize"`
	JobLogCommitTimeout   int      `json:"jobLogCommitTimeout"`
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
