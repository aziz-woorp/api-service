package config

import (
	"log/slog"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type (
	APIConfig struct {
		AppEnv          string        `default:"development" envconfig:"APP_ENV"`
		AppPort         string        `default:"8080"        envconfig:"APP_PORT"`
		AppName         string        `default:""            envconfig:"NEW_RELIC_APP_NAME"`
		LogLevel        slog.Level    `default:"INFO"        envconfig:"LOG_LEVEL"`
		TranslationsDir string        `default:"translation" envconfig:"APP_TRANSLATIONS_DIR"`
		MaxRetries      int           `default:"2"           envconfig:"MAX_RETRIES"`
		RetryWaitMin    time.Duration `default:"1s"          envconfig:"RETRY_WAIT_MIN"`
		RetryWaitMax    time.Duration `default:"2s"          envconfig:"RETRY_WAIT_MAX"`
	}

	apiOptions struct{}
)

type Func[T any] func(*T) error

func NewAPIConfig(opts ...Func[APIConfig]) (*APIConfig, error) {
	cfg := &APIConfig{}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// APIOptions contains all API options.
var APIOptions apiOptions

// LoadConfig implements option pattern for loading config from environment.
func (apiOptions) LoadConfig(prefix string) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		c := &APIConfig{}

		err := envconfig.Process(prefix, c)
		if err != nil {
			return err
		}

		var optFns []Func[APIConfig]

		optFns = append(optFns, APIOptions.AppEnv(c.AppEnv))
		optFns = append(optFns, APIOptions.AppPort(c.AppPort))
		optFns = append(optFns, APIOptions.AppName(c.AppName))
		optFns = append(optFns, APIOptions.LogLevel(c.LogLevel))
		optFns = append(optFns, APIOptions.TranslationsDir(c.TranslationsDir))
		optFns = append(optFns, APIOptions.MaxRetries(c.MaxRetries))
		optFns = append(optFns, APIOptions.RetryWaitMin(c.RetryWaitMin))
		optFns = append(optFns, APIOptions.RetryWaitMax(c.RetryWaitMax))

		for _, fn := range optFns {
			if err := fn(cfg); err != nil {
				return err
			}
		}

		return nil
	}
}

func (apiOptions) AppEnv(env string) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.AppEnv = env
		return nil
	}
}

func (apiOptions) AppPort(port string) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.AppPort = port
		return nil
	}
}

func (apiOptions) AppName(name string) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.AppName = name
		return nil
	}
}

func (apiOptions) LogLevel(level slog.Level) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.LogLevel = level
		return nil
	}
}

func (apiOptions) TranslationsDir(dir string) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.TranslationsDir = dir
		return nil
	}
}

func (apiOptions) MaxRetries(nums int) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.MaxRetries = nums
		return nil
	}
}

func (apiOptions) RetryWaitMin(d time.Duration) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.RetryWaitMin = d
		return nil
	}
}

func (apiOptions) RetryWaitMax(d time.Duration) Func[APIConfig] {
	return func(cfg *APIConfig) error {
		cfg.RetryWaitMax = d
		return nil
	}
}