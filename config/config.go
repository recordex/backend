package config

import "os"

type Config struct {
	CloudStorageBucketName string
	EthereumNodeURL        string
}

var config *Config

func init() {
	env := os.Getenv("APP_ENV")
	switch env {
	case "prd":
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
		}
	case "dev":
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
		}
	default:
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
		}
	}
}

func Get() *Config {
	return config
}
