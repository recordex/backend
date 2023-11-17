package config

import "os"

type Config struct {
	CloudStorageBucketName string
	EthereumNodeURL        string
	RecordContractAddress  string
}

var config *Config

func init() {
	env := os.Getenv("APP_ENV")
	switch env {
	case "prd":
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
			RecordContractAddress:  "0xC3e4bb03b22C7DcB3715A2f973f25Ba72d9A2e37",
		}
	case "dev":
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
			RecordContractAddress:  "0xC3e4bb03b22C7DcB3715A2f973f25Ba72d9A2e37",
		}
	default:
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
			RecordContractAddress:  "0xC3e4bb03b22C7DcB3715A2f973f25Ba72d9A2e37",
		}
	}
}

func Get() *Config {
	return config
}
