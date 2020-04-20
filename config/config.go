package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var (
	//Token is the discord bot token
	Token string
	//BotPrefix is the prefixed used prior to every bot command in discord
	BotPrefix string
	//GiphyKey is the key to the Giphy API
	GiphyKey string
	//RiotKey is the key to the Riot API
	RiotKey string

	config *configStruct
)

type configStruct struct {
	Token     string `json:"Token"`
	BotPrefix string `json:"BotPrefix"`
	GiphyKey  string `json:"GiphyAPIKey"`
	RiotKey   string `json:"RiotKey"`
}

//ReadConfig is a function to export hidden variables
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
	RiotKey = config.RiotKey

	return nil
}
