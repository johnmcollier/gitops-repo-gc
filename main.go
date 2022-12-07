package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

var appDataOrg = "redhat-appstudio-appdata"

func main() {
	if os.Getenv("GITHUB_TOKEN") == "" {
		fmt.Println("GITHUB_TOKEN must be set as an environment variable")
	}
	githubToken := os.Getenv("GITHUB_TOKEN")

	// Initialize an authenticated github client, along with an unauthenticated one
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	//client := github.NewClient(nil)
	authClient := github.NewClient(tc)

	// Parse command line flags to determine which operation to perform
	// Options are:
	// Delete all repositories owned by user: xyz
	// Delete all invalid repositories (ones starting with a dash)
	var operation, keyword, repo string
	flag.StringVar(&operation, "operation", "", "The operation to perform. One of: delete-by-user or delete-invalid")
	flag.StringVar(&keyword, "keyword", "", "The keyword(s) to match gitops repositories on")
	flag.StringVar(&repo, "repo", "", "The name of a repository")
	flag.Parse()

	// Check the values of the flags before proceding
	if operation == "" {
		log.Fatal("usage: --operation must be set as a command-line flag")
	}

	if operation != "delete-by-keyword" && operation != "delete-invalid" && operation != "list-all" && operation != "delete" {
		log.Fatal("usage: The only valid options for '--operation' are delete-by-keyword or delete-invalid")
	}

	if operation == "delete-by-keyword" {
		if keyword == "" {
			log.Fatal("usage: If deleting repositories by keyword, the '--keyword' flag must be set")
		}

		keywordRepos, err := searchReposByKeyword(ctx, authClient, keyword)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = deleteRepos(ctx, authClient, keywordRepos)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else if operation == "delete-invalid" {
		invalidRepos, err := listInvalidRepos(ctx, authClient)
		if err != nil {
			log.Fatal(err.Error())
		}
		err = deleteRepos(ctx, authClient, invalidRepos)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else if operation == "list-all" {
		allRepos, err := listAllRepos(ctx, authClient)
		if err != nil {
			log.Fatal(err.Error())
		}
		for _, repo := range allRepos {
			fmt.Println(*repo.Name)
		}

	} else {
		if repo == "" {
			log.Fatal("usage: --repo <repo-name> must be passed in as a flag when using the 'delete' operation")
		}
		err := deleteRepo(ctx, authClient, repo)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}

func listAllRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, appDataOrg, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}
	return allRepos, nil

}

func listInvalidRepos(ctx context.Context, client *github.Client) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	count := 0
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, appDataOrg, opt)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			repoName := *repo.Name

			if repoName[0:1] == "-" {
				allRepos = append(allRepos, repo)
			}
		}

		// ToDo: Cleanup
		// There's a lot (> 10k) of invalid repos right now. Limit to 1000 returned to avoid rate limiting the GitHub token
		count++
		if count == 40 {
			break
		}
		// ToDo: cleanup
		// By default it seems go-github returns the oldest repositories first, rather than newest. So after we get the first set of results,
		// navigate to the "last page" (the newest repositories) and move backwards
		// This is because most of the invalid repositories are front loaded (i.e. newer) so we don't want to waste API calls going through
		// pages of old, valid repositories
		if count == 1 {
			//fmt.Printf("Last page: %d\n", resp.LastPage)
			opt.Page = resp.LastPage
		} else {
			//fmt.Printf("Prev page: %d\n", resp.PrevPage)
			opt.Page = resp.PrevPage
		}

		if opt.Page == 0 {
			break
		}

	}
	return allRepos, nil
}

func searchReposByKeyword(ctx context.Context, client *github.Client, keyword string) ([]*github.Repository, error) {
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{PerPage: 100},
		TextMatch:   true,
	}
	var allRepos []*github.Repository
	query := "org:" + appDataOrg + " " + keyword
	for {
		searchResult, resp, err := client.Search.Repositories(ctx, query, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, searchResult.Repositories...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage

	}

	return allRepos, nil
}

func deleteRepos(ctx context.Context, client *github.Client, repos []*github.Repository) error {
	for _, repo := range repos {
		fmt.Println("Deleting repo: " + *repo.Name)
		_, err := client.Repositories.Delete(ctx, appDataOrg, *repo.Name)
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func deleteRepo(ctx context.Context, client *github.Client, repo string) error {
	fmt.Println("Deleting repo: " + repo)
	_, err := client.Repositories.Delete(ctx, appDataOrg, repo)
	if err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return nil
}
