package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

var (
	configJSONPath = "./config.json"
)

type JSONConfig struct {
	Outpath string `json:"outpath"`
	JSPath  string `json:"jspath"`
}

func IsVideoQuality(segment string) bool {
	re := regexp.MustCompile(`^\d+p\d*$`)
	return re.MatchString(segment)
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func CreateVideo(dir, filename string) string {
	re := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	filename = re.ReplaceAllString(filename, "_")
	filePath := filepath.Join(dir, fmt.Sprintf("%s.mp4", filename))
	counter := 1
	for FileExists(filePath) {
		filePath = filepath.Join(dir, fmt.Sprintf("%s (%v).mp4", filename, counter))
		counter++
	}
	_, err := os.Create(filePath)
	if err != nil {
		log.Fatal("Failted to create new video")
	}
	return filePath
}

// Open file and append bytes
func AppendToFile(filePath string, data []byte) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open the files: %w", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}
	return nil
}

func getFullURL(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	v, _ := path.Split(parsed.Path)
	fullURL := &url.URL{
		Scheme: "https",
		Host:   parsed.Host,
		Path:   v,
	}
	return fullURL.String()
}

func ConstructURL(urls []string, quality string) string {
	var u string
	if quality != "" {
		for _, x := range urls {
			if strings.Contains(x, quality) {
				u = getFullURL(x)
			}
		}
	} else {
		u = getFullURL(urls[0])
	}
	return u
}

func IsNotEmpty(v interface{}) bool {
	return reflect.ValueOf(v).NumField() > 0
}

func PrintConfig(cfg JSONConfig) error {
	cfgJson, err := ConvertConfigToIndentedJSON(cfg)
	if err != nil {
		return err
	}
	fmt.Println(string(cfgJson))
	return nil
}

func ConvertConfigToIndentedJSON(cfg JSONConfig) ([]byte, error) {
	b, err := readFile(configJSONPath)
	if err != nil {
		return nil, err
	}
	var jsonData map[string]string
	if err := json.Unmarshal(b, &jsonData); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}
	v := reflect.ValueOf(cfg)
	typeCfg := v.Type()
	for i := 0; i < v.NumField(); i++ {
		jsonTag := typeCfg.Field(i).Tag.Get("json")
		value := v.Field(i).Interface().(string)
		jsonData[jsonTag] = value
	}
	indented, err := json.MarshalIndent(jsonData, "", " ")
	if err != nil {
		return nil, fmt.Errorf("error marshalling JSON: %v", err)
	}
	return indented, nil
}

func OverwriteConfig(cfg JSONConfig) error {
	cfgJson, err := ConvertConfigToIndentedJSON(cfg)
	if err != nil {
		return err
	}
	if err = os.WriteFile(configJSONPath, cfgJson, 0644); err != nil {
		return err
	}
	return nil
}

func ReadConfig() (JSONConfig, error) {
	bytes, err := readFile(configJSONPath)
	if err != nil {
		return JSONConfig{}, err
	}
	var config JSONConfig
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return JSONConfig{}, err
	}
	return config, nil
}

func ArePathsValid(opts interface{}) error {
	v := reflect.ValueOf(opts)
	typeOfCfg := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := typeOfCfg.Field(i)
		value := v.Field(i).Interface().(string)
		if !IsPathValid(value) {
			return fmt.Errorf("the path is not valid: {%s: %s}", field.Name, value)
		}
	}
	return nil
}

func IsPathValid(filepath string) bool {
	f, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	defer f.Close()
	return true
}

func readFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open the filePath: %s. error: %s", filePath, err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the content: %s, error: %s", filePath, err)
	}
	return bytes, nil
}

func SetConfig(filePath string, v interface{}) error {
	if FileExists(filePath) {
		data, err := json.MarshalIndent(v, "", " ")
		if err != nil {
			return err
		}
		err = os.WriteFile(filePath, data, 0644)
		if err != nil {
			return err
		}
	}
	return fmt.Errorf("the filepath is incorrect: %s", filePath)
}
