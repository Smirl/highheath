package highheath

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v32/github"
)

func GithubClient() *github.Client {
	// Wrap the shared transport for use with the integration ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, 74377, 10821586, "high-heath-farm-cattery.2020-07-26.private-key.pem")
	if err != nil {
		log.Fatalf("Failed to create github client: %v", err)
	}

	// Use installation transport with client.
	return github.NewClient(&http.Client{Transport: itr})
}

func CreateComment(ctx context.Context, client *github.Client, comment *Comment) error {
	// Slug used for filename and branch name
	slug := comment.Date.Format("20060102-150405")

	// Get SHA of master branch
	ref, _, err := client.Git.GetRef(ctx, "smirl", "highheath", "heads/master")
	if err != nil {
		return err
	}

	// Create branch from lastest master SHA
	newRef := fmt.Sprintf("refs/heads/%s", slug)
	newRefObj := github.Reference{Ref: &newRef, Object: &github.GitObject{SHA: ref.Object.SHA}}
	if _, _, err := client.Git.CreateRef(ctx, "smirl", "highheath", &newRefObj); err != nil {
		return err
	}

	// Create file commit
	path := fmt.Sprintf("content/comments/%s.md", slug)
	message := fmt.Sprintf("🤖💬 New comment from %s", comment.GetName())
	opts := &github.RepositoryContentFileOptions{
		Content: comment.GetFileContent(),
		Branch:  &slug,
		Message: &message,
	}
	if _, _, err := client.Repositories.CreateFile(ctx, "smirl", "highheath", path, opts); err != nil {
		return err
	}

	// Create Pull Request
	base := "master"
	body := fmt.Sprintf("New comment:\n\n**Name**\n%s\n**Message**\n%s\n", comment.Name, comment.Message)
	newPR := github.NewPullRequest{Title: &message, Head: &slug, Base: &base, Body: &body}
	pr, _, err := client.PullRequests.Create(ctx, "smirl", "highheath", &newPR)
	if err != nil {
		return err
	}

	// Label the Pull Request
	labels := []string{"comment", "patch"}
	assignee := "smirl"
	updatedPR := github.IssueRequest{Labels: &labels, Assignee: &assignee}
	if _, _, err := client.Issues.Edit(ctx, "smirl", "highheath", *pr.Number, &updatedPR); err != nil {
		return err
	}
	return nil
}
