package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Env    string `yaml:"env"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Database struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
	Token struct {
		AccessTokenSecret  string `yaml:"accessTokenSecret"`
		RefreshTokenSecret string `yaml:"refreshTokenSecret"`
		AccessTokenExpire  int    `yaml:"accessTokenExpire"`
		RefreshTokenExpire int    `yaml:"refreshTokenExpire"`
	} `yaml:"token"`
}

func Load(file string) (Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed read config file %q", file)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed parsing yaml file %q", file)
	}
	return config, nil
}
