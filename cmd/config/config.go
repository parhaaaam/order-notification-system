package config

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/modules/config"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type PoolConfig struct {
	MaxConn int `mapstructure:"max-cons"`
	MinConn int `mapstructure:"min-cons"`
}

type RabbitMQConfig struct {
	ConnectionURL string `mapstructure:"connectionurl"`
}

type Config struct {
	Databases      config.DatabaseList `mapstructure:"databases,omitempty"`
	Pool           PoolConfig          `mapstructure:"pool"`
	RabbitMQClient RabbitMQConfig      `mapstructure:"rabbitmqclient"`
}

func Load() *Config {
	viper.SetDefault("databases.default.host", "localhost")
	viper.SetDefault("databases.default.port", "5432")
	viper.SetDefault("databases.default.user", "parham")
	viper.SetDefault("databases.default.pwd", "")
	viper.SetDefault("databases.default.name", "postgres")
	viper.SetDefault("databases.default.driver", "postgresql")
	viper.SetDefault("pool.max-cons", 3)
	viper.SetDefault("pool.min-cons", 1)
	viper.SetDefault("rabbitmqclient.connectionurl", "amqp://guest:guest@localhost:5672")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	for _, key := range viper.AllKeys() {
		_ = viper.BindEnv(key, strings.ToUpper(strings.ReplaceAll(key, ".", "_")))
		//cobra.CheckErr(err)
	}

	conf := Config{}
	if err := viper.Unmarshal(&conf); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}

	return &conf
}
