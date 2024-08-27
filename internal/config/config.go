package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Kostaaa1/twitchdl/internal/file"
	"github.com/Kostaaa1/twitchdl/types"
	"github.com/spf13/viper"
)

// config.json
func Get() (*types.JsonConfig, error) {
	var data types.JsonConfig

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(wd, "config.json")
	if !file.Exists(p) {
		f, err := os.Create(p)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		b, err := json.MarshalIndent(data, "", " ")
		if err != nil {
			return nil, err
		}
		if _, err := f.Write(b); err != nil {
			return nil, err
		}
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("json")
		viper.AddConfigPath(".")
		err := viper.ReadInConfig()
		if err != nil {
			return nil, err
		}
		viper.Unmarshal(&data)
	}
	return &data, nil
}
