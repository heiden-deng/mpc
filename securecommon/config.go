package common

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	log "github.com/sirupsen/logrus"
)

type Host struct {
	Address string `yaml:"address"`
	Port string `yaml:"port"`
	ID   string `yaml:"id"`
}

type Config struct {
	Controller struct{
		Address string `yaml:"address"`
		Port string `yaml:"port"`
	}`yaml:"controller"`
	Org1 Host `yaml:"org1"`
	Org2 Host `yaml:"org2"`
	Org3 Host `yaml:"org3"`
	Id string `yaml:"id"`
}

var Conf *Config

func init(){
	Conf = loadConfig()
}

func loadConfig() *Config{
	configContent,err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Panic("Read config err")
	}
	var conf Config
	err = yaml.Unmarshal(configContent, &conf)
	if err != nil {
		log.Panic("Cannot parse config file")
	}
	log.Info("Read conf finish")
	return &conf
}

