package helpers

import (
	"io/ioutil"
)

var config string

func GetConfigPort() string {
	if len(config) > 0 {
		return config
	}

	content, err := ioutil.ReadFile("config.txt")
	if err != nil {
		return "8888"
	}

	config = string(content)
	return config
}
