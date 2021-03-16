package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

func main() {
	if os.Getenv("AWS_SM_ID") == "" || os.Getenv("AWS_SM_REGION") == "" {
		log.Println("aws-env running locally, without AWS_SM_ID and AWS_SM_REGION")
		return
	}
	secretId := os.Getenv("AWS_SM_ID")
	region := os.Getenv("AWS_SM_REGION")

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}

	client := secretsmanager.NewFromConfig(cfg)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretId),
	}
	result, err := client.GetSecretValue(context.TODO(), input)

	if err != nil {
		log.Fatalf("failed to get secret values, %v", err)
	}

	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			log.Panic("Base64 Decode Error:", err)
			return
		}
		secretString = string(decodedBinarySecretBytes[:len])
	}

	var envars map[string]interface{}
	json.Unmarshal([]byte(secretString), &envars)

	for key, value := range envars {
		fmt.Printf("export %s=$'%s'\n", key, value)
	}
}
