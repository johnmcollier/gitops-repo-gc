# gitops-repo-gc

A simple command-line utility to garbage collect invalid or no longer needed repositories from the GitOps repository that HAS uses.

## Usage:

The tool requires that the `GITHUB_TOKEN` environment variable be exported before using it. The token that you use, must have sufficient permissions to delete repositories from the specified org. 

To delete invalid repositories in the GitOps org (up to 500 at a time) run:
```
./gitops-repo-gc --operation delete-invalid
```

To delete repositories by a given keyword run:
```
./gitops-repo-gc --operation delete-by-keyword --keyword some-keyword
```
