package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Config struct {
	Mongo struct {
		Example struct {
			Uri      string `yaml:"uri" json:"-"`
			Database string `yaml:"database" json:"-"`
		} `yaml:"example" json:"-"`
	} `yaml:"mongo" json:"-"`
	Log struct {
		Filename string `yaml:"filename" json:"-"`
		LongFile bool   `yaml:"long_file" json:"-"`
		Level    string `yaml:"level" json:"-"`
	} `yaml:"log" json:"-"`
	Youtube struct {
		Key string `yaml:"key" json:"-"`
	} `yaml:"youtube" json:"-"`
	Proxy struct {
		Example string `yaml:"example" json:"-"`
	} `yaml:"proxy" json:"-"`
}

func (c *Config) LoadConfig(configPath string) (err error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Panicln(err)
	}

	err = yaml.Unmarshal(configData, c)
	if err != nil {
		log.Panicln(err)
	}

	return
}

func NewConfig() (config *Config, err error) {
	configFilePtr := flag.String("c", "", "config file")

	flag.Parse()

	config = &Config{}
	configFile := *configFilePtr
	if configFile != "" {
		err = config.LoadConfig(configFile)
		if err != nil {
			log.Panicln(err)
		}
	}

	return
}
