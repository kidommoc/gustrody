package config

type Config struct {
	Debug   bool   `json:"debug"`
	Site    string `json:"site"`
	Port    int    `json:"port"`
	HmacKey string `json:"hmacKey"`

	// logging
	Logfile  string `json:"logfile"`
	LogSplit int    `json:"logSplit"`
	LogLevel int    `json:"logLevel"`

	// database
	PqUser   string `json:"pqUser"`
	PqSecret string `json:"pqSecret"`
	RdSecret string `json:"rdSecret"`

	// static files
	ImgDir string `json:"imgDir"`

	// perference
	MaxContentLength int `json:"maxCotentLength"`
	MaxImgInPost     int `json:"maxImgInPost"`
}

var config *Config

func Get() Config {
	if config == nil {
		loadEnv()
	}
	return *config
}
