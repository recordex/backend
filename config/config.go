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
			RecordContractAddress:  "0x590b51f9D972625B263eAD11417832Fcf4fc724c",
		}
	case "dev":
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
			RecordContractAddress:  "0x590b51f9D972625B263eAD11417832Fcf4fc724c",
		}
	default:
		config = &Config{
			CloudStorageBucketName: "recordex",
			EthereumNodeURL:        "https://sepolia.infura.io/v3",
			RecordContractAddress:  "0x590b51f9D972625B263eAD11417832Fcf4fc724c",
		}
	}
}

func Get() *Config {
	return config
}
