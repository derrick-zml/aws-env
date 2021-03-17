package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
	region := ""
	if os.Getenv("AWS_REGION") != "" {
		region = os.Getenv("AWS_SM_REGION")
	} else if os.Getenv("AWS_SM_REGION") != "" {
		region = os.Getenv("AWS_SM_REGION")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("failed to load SDK configuration, %v", err)
	}

	if os.Getenv("AWS_SM_ID") != "" {
		secretId := os.Getenv("AWS_SM_ID")
		client := secretsmanager.NewFromConfig(cfg)
		ExportVariablesFromSecretsManager(client, secretId)
	} else if os.Getenv("AWS_ENV_PATH") != "" {
		path := os.Getenv("AWS_ENV_PATH")
		client := ssm.NewFromConfig(cfg)
		ExportVariablesFromSSM(client, path, true, "")
	} else {
		log.Println("aws-env running locally")
		return
	}
}

func ExportVariablesFromSecretsManager(client *secretsmanager.Client, secretId string) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretId),
	}
	result, err := client.GetSecretValue(context.TODO(), input)

	if err != nil {
		log.Fatalf("failed to get secret values, %v", err)
	}

	if result.SecretString != nil {
		secretString := *result.SecretString
		var envars map[string]interface{}
		json.Unmarshal([]byte(secretString), &envars)

		for key, value := range envars {
			fmt.Printf("export %s=$'%s'\n", key, value)
		}
	}
}

func ExportVariablesFromSSM(client *ssm.Client, path string, recursive bool, nextToken string) {
	input := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: true,
		Recursive:      recursive,
	}

	if nextToken != "" {
		input.NextToken = aws.String(nextToken)
	}

	result, err := client.GetParametersByPath(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to get paremters by path, %v", err)
	}

	for _, parameter := range result.Parameters {
		name := *parameter.Name
		value := *parameter.Value

		env := strings.Replace(strings.Trim(name[len(path):], "/"), "/", "_", -1)
		value = strings.Replace(value, "\n", "\\n", -1)

		fmt.Printf("export %s=$'%s'\n", env, value)
	}

	if result.NextToken != nil {
		ExportVariablesFromSSM(client, path, recursive, *result.NextToken)
	}
}
