package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kostaaa1/twitchdl/types"
	"github.com/spf13/viper"
)

func Get() (*types.JsonConfig, error) {
	var data types.JsonConfig
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	p := filepath.Join(wd, "config.json")

	_, err = os.Stat(p)
	if os.IsNotExist(err) {
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

func ValidateCreds() error {
	cfg, err := Get()
	if err != nil {
		return err
	}
	errors := []string{}
	if cfg.Creds.AccessToken == "" {
		errors = append(errors, "AccessToken")
	}
	if cfg.Creds.ClientSecret == "" {
		errors = append(errors, "ClientSecret")
	}
	if cfg.Creds.ClientID == "" {
		errors = append(errors, "ClientID")
	}
	if len(errors) > 0 {
		for _, err := range errors {
			msg := fmt.Sprintf("missing %s from config.json", err)
			return fmt.Errorf(msg)
		}
	}
	return nil
}
