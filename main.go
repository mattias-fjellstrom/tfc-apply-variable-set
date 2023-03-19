package main

import (
	"context"
	"flag"
	"log"
	"os"

	tfe "github.com/hashicorp/go-tfe"
)

const (
	ENV_TERRAFORM_CLOUD_TOKEN        = "TERRAFORM_CLOUD_TOKEN"
	ENV_TERRAFORM_CLOUD_ORGANIZATION = "TERRAFORM_CLOUD_ORGANIZATION"
	ENV_TERRAFORM_CLOUD_WORKSPACE    = "TERRAFORM_CLOUD_WORKSPACE"
)

var organizationName string
var workspaceName string
var variableSetName string

func init() {
	flag.StringVar(&organizationName, "organization", "", "Terraform Cloud organization name")
	flag.StringVar(&workspaceName, "workspace", "", "Desired Terraform Cloud workspace name")
	flag.StringVar(&variableSetName, "varset", "", "Variable set name")
}

func main() {
	flag.Parse()

	if organizationName == "" {
		log.Println("No organization name provided as input argument, will fall back to environment variable")
		_, ok := os.LookupEnv(ENV_TERRAFORM_CLOUD_ORGANIZATION)
		if !ok {
			log.Fatalf("The organization name must be provided either as an input parameter or in the %s environment variable", ENV_TERRAFORM_CLOUD_ORGANIZATION)
		}
		organizationName = os.Getenv(ENV_TERRAFORM_CLOUD_ORGANIZATION)
		log.Println("Organization name read from environment variable")
	}

	if workspaceName == "" {
		log.Println("No workspace name provided as input argument, will fall back to environment variable")
		_, ok := os.LookupEnv(ENV_TERRAFORM_CLOUD_WORKSPACE)
		if !ok {
			log.Fatalf("A workspace name must be provided either as an input parameter or in the %s environment variable", ENV_TERRAFORM_CLOUD_WORKSPACE)
		}
		workspaceName = os.Getenv(ENV_TERRAFORM_CLOUD_WORKSPACE)
		log.Println("Workspace name read from environment variable")
	}

	if variableSetName == "" {
		log.Fatal("Variable set name not provided as an input parameter")
	}

	token, ok := os.LookupEnv(ENV_TERRAFORM_CLOUD_TOKEN)
	if !ok || token == "" {
		log.Fatalf("%s environment variable must be set with a valid token", ENV_TERRAFORM_CLOUD_TOKEN)
	}

	config := &tfe.Config{
		Token:             token,
		RetryServerErrors: true,
	}

	client, err := tfe.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	workspace, err := client.Workspaces.Read(ctx, organizationName, workspaceName)
	if err != nil {
		log.Fatal(err)
	}

	varSets, err := client.VariableSets.List(ctx, organizationName, &tfe.VariableSetListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	for _, varSet := range varSets.Items {
		if varSet.Name == variableSetName {
			err = client.VariableSets.ApplyToWorkspaces(ctx, varSet.ID, &tfe.VariableSetApplyToWorkspacesOptions{
				Workspaces: []*tfe.Workspace{
					workspace,
				},
			})
			if err != nil {
				log.Fatal(err)
			}
			break
		}
	}
}
