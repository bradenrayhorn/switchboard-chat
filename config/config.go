package config

import (
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var RsaPublic *rsa.PublicKey

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		log.Println(err.Error())
	}

	viper.SetDefault("token_expiration", time.Hour*24)
	viper.SetDefault("rsa_path", "jwt_rsa")

	_ = viper.BindEnv("mongo_host", "MONGO_HOST")
	_ = viper.BindEnv("mongo_port", "MONGO_PORT")
	_ = viper.BindEnv("mongo_username", "MONGO_USERNAME")
	_ = viper.BindEnv("mongo_password", "MONGO_PASSWORD")
	_ = viper.BindEnv("mongo_database", "MONGO_DATABASE")
	_ = viper.BindEnv("rsa_path", "RSA_PATH")

	loadRsaKey()
}

func loadRsaKey() {
	publicKey, err := readKey(true)
	if err != nil {
		log.Fatalf("failed to load public rsa key: %s", err)
	}
	RsaPublic = publicKey.(*rsa.PublicKey)
}

func readKey(public bool) (interface{}, error) {
	filePath := viper.GetString("rsa_path")
	if public {
		filePath += ".pub"
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	keyBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var rsaKey interface{}
	if public {
		rsaKey, err = jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	}
	return rsaKey, err
}
