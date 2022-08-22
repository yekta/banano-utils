package sharedUtils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func GetEnv(key string) string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("\nNo .env file, will try to use env variables...")
	}
	return os.Getenv(key)
}
