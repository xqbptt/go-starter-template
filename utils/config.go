package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	PORT        string           `mapstructure:"PORT"`
	DB_URL      string           `mapstructure:"DB_URL"`
	SCHEDULER   SchedulerConfig  `mapstructure:"SCHEDULER"`
	RENDERER    RendererConfig   `mapstructure:"RENDERER"`
	AWS         AwsConfig        `mapstructure:"AWS"`
	INSTAGRAM   InstagramConfig  `mapstructure:"INSTAGRAM"`
	LLM         LLMConfig        `mapstructure:"LLM"`
	IMAGES      ImagesConfig     `mapstructure:"IMAGES"`
	OCI_STORAGE OciStorageConfig `mapstructure:"OCI_STORAGE"`
	CORS        []string         `mapstructure:"CORS"`
	TOKEN       TokenConfig      `mapstructure:"TOKEN"`
	GOOGLE      GoogleConfig     `mapstructure:"GOOGLE"`
}

type SchedulerConfig struct {
	SCHEDULER_ENDPOINT string `mapstructure:"SCHEDULER_ENDPOINT"`
	SCHEDULER_PASSWORD string `mapstructure:"SCHEDULER_PASSWORD"`
	SCHEDULER_USERNAME string `mapstructure:"SCHEDULER_USERNAME"`
}

type RendererConfig struct {
	RENDERER_ENDPOINT string `mapstructure:"RENDERER_ENDPOINT"`
	RENDERER_PASSWORD string `mapstructure:"RENDERER_PASSWORD"`
	RENDERER_USERNAME string `mapstructure:"RENDERER_USERNAME"`
}

type AwsConfig struct {
	ACCESS_KEY_ID     string `mapstructure:"ACCESS_KEY_ID"`
	ACCESS_KEY_SECRET string `mapstructure:"ACCESS_KEY_SECRET"`
	REGION            string `mapstructure:"REGION"`
}

type InstagramConfig struct {
	INSTAGRAM_CLIENT_ID            string `mapstructure:"INSTAGRAM_CLIENT_ID"`
	INSTAGRAM_CLIENT_SECRET        string `mapstructure:"INSTAGRAM_CLIENT_SECRET"`
	INSTAGRAM_REDIRECT_URI         string `mapstructure:"INSTAGRAM_REDIRECT_URI"`
	INSTAGRAM_WEBHOOK_VERIFY_TOKEN string `mapstructure:"INSTAGRAM_WEBHOOK_VERIFY_TOKEN"`
}

type GoogleConfig struct {
	CLIENT_ID     string `mapstructure:"CLIENT_ID"`
	CLIENT_SECRET string `mapstructure:"CLIENT_SECRET"`
	REDIRECT_URI  string `mapstructure:"REDIRECT_URI"`
}

type LLMConfig struct {
	GEMINI_API_KEY string `mapstructure:"GEMINI_API_KEY"`
}

type ImagesConfig struct {
	UNSPLASH_CLIENT_ID             string `mapstructure:"UNSPLASH_CLIENT_ID"`
	BING_SEARCH_API_KEY            string `mapstructure:"BING_SEARCH_API_KEY"`
	GOOGLE_CUSTOM_SEARCH_API_KEY   string `mapstructure:"GOOGLE_CUSTOM_SEARCH_API_KEY"`
	GOOGLE_CUSTOM_SEARCH_ENGINE_ID string `mapstructure:"GOOGLE_CUSTOM_SEARCH_ENGINE_ID"`
}

type TokenConfig struct {
	ACCESS_SECRET_KEY      string        `mapstructure:"ACCESS_SECRET_KEY"`
	ACCESS_PUBLIC_KEY      string        `mapstructure:"ACCESS_PUBLIC_KEY"`
	ACCESS_TOKEN_DURATION  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	REFRESH_SECRET_KEY     string        `mapstructure:"REFRESH_SECRET_KEY"`
	REFRESH_PUBLIC_KEY     string        `mapstructure:"REFRESH_PUBLIC_KEY"`
	REFRESH_TOKEN_DURATION time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
}

type OciStorageConfig struct {
	HOST           string `mapstructure:"HOST"`
	KEY_ID         string `mapstructure:"KEY_ID"`
	NAMESPACE      string `mapstructure:"NAMESPACE"`
	COMPARTMENT_ID string `mapstructure:"COMPARTMENT_ID"`
	BUCKET_NAME    string `mapstructure:"BUCKET_NAME"`
	PRIVATE_KEY    string `mapstructure:"PRIVATE_KEY"`
	PAR_PREFIX     string `mapstructure:"PAR_PREFIX"`
}

var (
	cfg *Config
)

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (config Config, err error) {
	if cfg != nil {
		config = *cfg
		return
	}

	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("toml")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return
	}
	config = *cfg
	return
}
