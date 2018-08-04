package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
)

func mergePullRequest(owner, repo string, number int) (*string, int, error) {
	ctx := context.Background()
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	client := github.NewClient(oauth2.NewClient(ctx, src))

	commit, resp, err := client.Repositories.Merge(
		context.Background(), owner, repo, number, nil, nil,
	)
	if err != nil {
		return commit.URL, resp.StatusCode, err
	}
	return commit.URL, resp.StatusCode, nil
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	prEvent := new(github.PullRequestEvent)
	if err := json.Unmarshal([]byte(request.Body), &prEvent); err != nil {
		return events.APIGatewayProxyResponse{}, errors.New("Failed to unmarshal JSON")
	}

	owner := prEvent.GetRepo().Owner.Login
	repo := prEvent.GetRepo().Name
	number := prEvent.GetNumber()
	commitUrl, statusCode, err := mergePullRequest(owner, repo, number)
	if err != nil {
		// Notify failed result to slack
		return events.APIGatewayProxyResponse{statusCode: statusCode}, err
	}

	// Notify successful result to slack
	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
