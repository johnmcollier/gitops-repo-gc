#!/bin/sh

# Install Kubectl
cd /tmp
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# Get the list of all Application resources
kubectl get applications --all-namespaces > all-apps.yaml

# Get the list of all GitOps repos
/gitops-repo-gc --operation list-all > all-repos.txt

# Determine which gitops repositories need to be cleaned up
touch orphaned-repos.txt
while read p; do
    cat all-apps.yaml | grep $p > /dev/null
    if [ $? -ne 0 ]; then
        echo $p >> orphaned-repos.txt
    fi
done <all-repos.txt

# Delete the orphaned gitops repositories
while read p; do
    /gitops-repo-gc --operation delete --repo $p
    if [ $? -ne 0 ]; then
        exit 1
    fi
done <orphaned-repos.txt