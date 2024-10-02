package config

type Config struct {
	Debug bool `json:"debug"`

	Scheme  string `json:"scheme"`
	Domain  string `json:"domain"`
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
	MaxContentLength int  `json:"maxCotentLength"`
	MaxImgInPost     int  `json:"maxImgInPost"`
	AllowRegistry    bool `json:"allowRegistry"`
}

var config *Config

func Get() Config {
	if config == nil {
		loadEnv()
	}
	return *config
}
