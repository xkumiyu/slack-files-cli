package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

// Config is configuration structure
type Config struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Token    string `json:"token"`
}

// ConfigFile is path of config file
var ConfigFile = ".slack-files-cli.json"

// ReadConfig is function to read config from json file
func ReadConfig() *Config {
	bytes, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	cfg := new(Config)
	if err := json.Unmarshal(bytes, cfg); err != nil {
		log.Fatal(err)
	}

	return cfg
}

// Exists is function that return True if path refers to an existing path
func Exists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}
