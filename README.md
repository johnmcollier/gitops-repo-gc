# gitops-repo-gc

A simple command-line utility to garbage collect invalid or no longer needed repositories from the GitOps repository that HAS uses.

## Usage:

The tool requires that the `GITHUB_TOKEN` environment variable be exported before using it. The token that you use, must have sufficient permissions to delete repositories from the specified org. 

To delete invalid repositories in the GitOps org (up to 300 at a time) run:
```
./has-repo-gc --operation delete-valid
```

To delete repositories by a given keyword run:
```
./has-repo-gc --operation delete-by-keyword --keyword some-keyword
```