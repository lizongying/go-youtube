package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Config struct {
	Mongo struct {
		Uri                   string `yaml:"uri" json:"-"`
		Database              string `yaml:"database" json:"-"`
		CollectionYoutubeUser string `yaml:"collection_youtube_user" json:"-"`
	} `yaml:"mongo" json:"-"`
	Log struct {
		Filename string `yaml:"filename" json:"-"`
		LongFile bool   `yaml:"long_file" json:"-"`
		Level    string `yaml:"level" json:"-"`
	} `yaml:"log" json:"-"`
	Youtube struct {
		Key string `yaml:"key" json:"-"`
	} `yaml:"youtube" json:"-"`
}

func (c *Config) LoadConfig(configPath string) (err error) {
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Panicln(err)
		return
	}
	err = yaml.Unmarshal(configData, c)
	return
}

func NewConfig() (config *Config, err error) {
	configFilePtr := flag.String("c", "", "config file")

	mongoUriPtr := flag.String("mongo-uri", "", "")
	mongoDatabasePtr := flag.String("mongo-database", "", "")
	mongoCollectionYoutubeUserPtr := flag.String("mongo-collection-youtube-user", "", "")
	logFilenamePtr := flag.String("log-filename", "", "")
	logLongFilePtr := flag.Bool("log-long-file", false, "")
	logLevelPtr := flag.String("log-level", "", "")
	youtubeKeyPtr := flag.String("youtube-key", "", "")
	flag.Parse()

	config = &Config{}
	configFile := *configFilePtr
	if configFile != "" {
		err = config.LoadConfig(configFile)
		if err != nil {
			log.Panicln(err)
		}
	}

	if config.Mongo.Uri == "" {
		config.Mongo.Uri = *mongoUriPtr
	}
	if config.Mongo.Database == "" {
		config.Mongo.Database = *mongoDatabasePtr
	}
	if config.Mongo.CollectionYoutubeUser == "" {
		config.Mongo.CollectionYoutubeUser = *mongoCollectionYoutubeUserPtr
	}
	if config.Log.Filename == "" {
		config.Log.Filename = *logFilenamePtr
	}
	if !config.Log.LongFile {
		config.Log.LongFile = *logLongFilePtr
	}
	if config.Log.Level == "" {
		config.Log.Level = *logLevelPtr
	}
	if config.Youtube.Key == "" {
		config.Youtube.Key = *youtubeKeyPtr
	}

	return
}
