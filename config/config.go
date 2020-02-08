package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	Token     string
	BotPrefix string
	GiphyKey  string

	config *configStruct
)

type configStruct struct {
	Token     string `json:"Token"`
	BotPrefix string `json:"BotPrefix"`
	GiphyKey  string `json:"GiphyAPIKey"`
}

func ReadConfig() error {
	fmt.Println("Reading from config file...")

	file, err := ioutil.ReadFile("./config/config.json")

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	//	fmt.Println(string(file))

	err = json.Unmarshal(file, &config)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	Token = config.Token
	BotPrefix = config.BotPrefix
	GiphyKey = config.GiphyKey

	return nil
}
