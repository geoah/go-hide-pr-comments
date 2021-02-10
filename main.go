package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/google/go-github/v33/github"
	"github.com/sethvargo/go-githubactions"
	"golang.org/x/oauth2"
)

func main() {
	prNumberString := githubactions.GetInput("pr_number")
	if prNumberString == "" {
		githubactions.Fatalf("missing input 'pr_number'")
	}

	githubToken := githubactions.GetInput("github_token")
	if githubToken == "" {
		githubactions.Fatalf("missing input 'github_token'")
	}
	githubactions.AddMask(githubToken)

	hideUserName := githubactions.GetInput("hide_user_name")
	if hideUserName == "" {
		githubactions.Fatalf("missing input 'hide_user_name'")
	}

	hideReason := githubactions.GetInput("hide_reason")
	if hideReason == "" {
		githubactions.Fatalf("missing input 'hide_reason'")
	}

	githubEventPath := os.Getenv("GITHUB_EVENT_PATH")
	if githubEventPath == "" {
		githubactions.Fatalf("$GITHUB_EVENT_PATH is empty")
		return
	}

	githubOwnerRepository := strings.Split(
		os.Getenv("GITHUB_REPOSITORY"),
		"/",
	)
	githubOwner := githubOwnerRepository[0]
	githubRepository := githubOwnerRepository[1]

	eventPathFile, err := os.Open(githubEventPath)
	defer eventPathFile.Close()
	if err != nil {
		githubactions.Fatalf(err.Error())
		return
	}

	prNumber, _ := strconv.Atoi(prNumberString)

	fmt.Println("githubToken=", githubToken)
	fmt.Println("hideUserName=", hideUserName)
	fmt.Println("hideReason=", hideReason)
	fmt.Println("githubOwner=", githubOwner)
	fmt.Println("githubRepository=", githubRepository)
	// fmt.Println("eventPathFile=", githubEventPath)

	// data := &pullRequestEventData{}

	// byteValue, err := ioutil.ReadAll(eventPathFile)
	// if err != nil {
	// 	githubactions.Fatalf("unable to read event file, err=%s", err.Error())
	// 	return
	// }

	// fmt.Println(string(byteValue))

	// if err := json.Unmarshal(byteValue, data); err != nil {
	// 	githubactions.Fatalf(err.Error())
	// 	return
	// }

	if prNumber == 0 {
		githubactions.Fatalf("missing pull request number")
		return
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: githubToken,
		},
	)

	fmt.Println("pullRequestID=", prNumber)
	fmt.Println("env=")
	spew.Dump(os.Environ())

	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// issue comments

	issueComments, _, err := client.Issues.ListComments(
		ctx,
		githubOwner,
		githubRepository,
		prNumber,
		&github.IssueListCommentsOptions{},
	)
	if err != nil {
		githubactions.Fatalf("error getting pr comments, err=%s", err.Error())
		return
	}

	fmt.Printf("Got %d comments on PR\n", len(issueComments))

	for _, issueComment := range issueComments {
		fmt.Println("Found comment", *issueComment.ID)
		spew.Dump(issueComment)
		if issueComment.User.Login == nil {
			continue
		}
		if *issueComment.User.Login != hideUserName {
			return
		}
		fmt.Println("Removing comment from issue", *issueComment.ID)
		if _, err := client.Issues.DeleteComment(
			ctx,
			githubOwner,
			githubRepository,
			*issueComment.ID,
		); err != nil {
			githubactions.Fatalf("error removing comment, err=%s", err.Error())
			return
		}
	}

	// pr comments

	prComments, _, err := client.PullRequests.ListComments(
		ctx,
		githubOwner,
		githubRepository,
		prNumber,
		&github.PullRequestListCommentsOptions{},
	)
	if err != nil {
		githubactions.Fatalf("error getting pr comments, err=%s", err.Error())
		return
	}

	fmt.Printf("Got %d comments on pr\n", len(prComments))

	for _, prComment := range prComments {
		fmt.Println("Found comment on pr", *prComment.ID)
		spew.Dump(prComment)
		if prComment.User.Login == nil {
			continue
		}
		if *prComment.User.Login != hideUserName {
			return
		}
		fmt.Println("Removing comment from pr", *prComment.ID)
		if _, err := client.Issues.DeleteComment(
			ctx,
			githubOwner,
			githubRepository,
			*prComment.ID,
		); err != nil {
			githubactions.Fatalf("error removing comment from pr, err=%s", err.Error())
			return
		}
	}

	// review comments

	reviews, _, err := client.PullRequests.ListReviews(
		ctx,
		githubOwner,
		githubRepository,
		prNumber,
		&github.ListOptions{},
	)

	fmt.Printf("Got %d reviews on PR\n", len(reviews))

	for _, review := range reviews {
		fmt.Printf("Got %d reviews on PR\n", len(reviews))

		reviewComments, _, err := client.PullRequests.ListReviewComments(
			ctx,
			githubOwner,
			githubRepository,
			prNumber,
			*review.ID,
			&github.ListOptions{},
		)
		if err != nil {
			githubactions.Fatalf("error getting review comments, err=%s", err.Error())
			return
		}

		fmt.Printf("Got %d comments on review\n", len(reviewComments))

		for _, reviewComment := range reviewComments {
			fmt.Println("Found comment", *reviewComment.ID)
			spew.Dump(reviewComment)
			if reviewComment.User.Login == nil {
				continue
			}
			if *reviewComment.User.Login != hideUserName {
				return
			}
			fmt.Println("Removing comment from review", *reviewComment.ID)
			if _, err := client.PullRequests.DeleteComment(
				ctx,
				githubOwner,
				githubRepository,
				*reviewComment.ID,
			); err != nil {
				githubactions.Fatalf("error removing comment from review, err=%s", err.Error())
				return
			}
		}
	}
}

type pullRequestEventData struct {
	Action *string `json:"action,omitempty"`
	Number *int    `json:"number,omitempty"`
}
