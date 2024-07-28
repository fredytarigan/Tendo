package config

import (
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName			string			`mapstructure:"APP_NAME"`
	AppMode			string 			`mapstructure:"APP_MODE"`
	AppHost			string 			`mapstructure:"APP_HOST"`
	AppPort			string 			`mapstructure:"APP_PORT"`

	WatchInterval		time.Duration	`mapstructure:"watchInterval"`
	WatchTargets  	 	[]WatchConfig	 `mapstructure:"watchTargets"`
	
}

type WatchConfig struct {
	SecretName		 		   string 	       			   `mapstructure:"secretName"`
	OpaqueSecretName		   string	   				   `mapstructure:"opaqueSecretName"`
	SecretNamespace			   string 	   				   `mapstructure:"secretNamespace"`
	CertificateID 	 	 		string						`mapstructure:"certificateID"`
	CertificateName		 		string 						`mapstructure:"certificateName"`
	CertificateRegion	 		string						`mapstructure:"certificateRegion"`
	CertificateResourceTypes	[]CertificateResourceType	 `mapstructure:"certificateResourceTypes"`
}

type CertificateResourceType struct {
	Name	string 		`mapstructure:"name"`
	Regions	[]string 	`mapstructure:"regions"`
}

func SetConfigFile(path string) {
	var (
		errConfig error
	)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	if errConfig = viper.ReadInConfig(); errConfig != nil {
		log.Fatal("error reading config file", errConfig)
		return
	}
}

// Load configuration from environment variables
func LoadConfig() Config {
	var (
		errConfig error
	)

	appName := parseEnv("APP_NAME", "Tendo")
	appMode := parseEnv("APP_MODE", "Development")
	appHost := parseEnv("APP_HOST", "127.0.0.1")
	appPort := parseEnv("APP_PORT", "8085")

	conf  := &Config {
		AppName: appName,
		AppMode: appMode,
		AppHost: appHost,
		AppPort: appPort,
	}

	if errConfig = viper.ReadInConfig(); errConfig != nil {
		log.Fatal("error reading config file", errConfig)
	}

	errConfig = viper.Unmarshal(&conf)
	if errConfig != nil {
		log.Fatal("error unmarshal config file", errConfig)
	}

	return *conf
}


func parseEnv(env string, defEnv string) string {
	if os.Getenv(env) == "" {
		return defEnv
	} else {
		return os.Getenv(env)
	}
}