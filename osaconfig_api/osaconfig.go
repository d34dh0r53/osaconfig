// osaconfig - delivers a curated osa config
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	uuid "github.com/satori/go.uuid"
	git "gopkg.in/libgit2/git2go.v27"
)

// SHA OSA SHA Given
var SHA string

// Blob structure
type Blob struct {
	*git.Blob
	id *git.Oid
}

// CheckOutOSA checks out the openstack-ansible repository at a given branch and SHA
func CheckOutOSA(sha string, branchName string) uuid.UUID {

	// Generate a UUID for us to place this particular clone
	// TODO(d34dh0r53): probably a better way to do this.
	u1 := uuid.Must(uuid.NewV4())

	//TODO(d34dh0r53): Make this configurable
	repoPath := flag.String("repo", "/Users/cdw/temp/OSA/"+u1.String(), "path to the git repository")
	osaURL := flag.String("osaurl", "https://github.com/openstack/openstack-ansible", "default url for OSA")
	flag.Parse()

	// Clone the repository into our repoPath/UUID
	repo, err := git.Clone(*osaURL, *repoPath, &git.CloneOptions{})
	if err != nil {
		log.Fatalf("Failed to clone repository: %s", err)
	}

	// Set git checkout options
	checkoutOpts := &git.CheckoutOpts{
		Strategy: git.CheckoutSafe | git.CheckoutRecreateMissing | git.CheckoutAllowConflicts | git.CheckoutUseTheirs,
	}

	// Create a new OID based on the SHA we would like to checkout
	myOID, err := git.NewOid(sha)
	if err != nil {
		log.Fatalf("Failed to create OID: %s", err)
	}

	// Getting the reference for the remote branch
	remoteBranch, err := repo.LookupBranch("origin/"+branchName, git.BranchRemote)
	if err != nil {
		log.Printf("Failed to find remote branch: %s", branchName)
		log.Fatal(err)
	}
	defer remoteBranch.Free()

	// Lookup the commit from remote branch
	commit, err := repo.LookupCommit(myOID)
	if err != nil {
		log.Printf("Failed to find remote branch commit: %s", branchName)
		log.Fatal(err)
	}

	// DEBUG
	spew.Dump(commit.Id())
	spew.Dump(commit.Author())
	spew.Dump(commit.Message())
	// END DEBUG

	localBranch, err := repo.LookupBranch(branchName, git.BranchLocal)
	// No local branch, lets create one
	if localBranch == nil || err != nil {
		// Creating local branch
		localBranch, err = repo.CreateBranch(branchName, commit, false)
		if err != nil {
			log.Printf("Failed to create local branch: %s", branchName)
			log.Fatal(err)
		}

		// Setting upstream to origin branch
		err = localBranch.SetUpstream("origin/" + branchName)
		if err != nil {
			log.Printf("Failed to create upstream to origin/%s", branchName)
			log.Fatal(err)
		}
	}
	if localBranch == nil {
		log.Fatal("Error while locating/creating local branch")
	}
	defer localBranch.Free()

	// Getting the tree for the branch
	localCommit, err := repo.LookupCommit(localBranch.Target())
	if err != nil {
		log.Printf("Failed to lookup for commit in local branch %s", branchName)
		log.Fatal(err)
	}
	defer localCommit.Free()

	tree, err := repo.LookupTree(localCommit.TreeId())
	if err != nil {
		log.Printf("Failed to lookup for tree %s", branchName)
		log.Fatal(err)
	}
	defer tree.Free()

	// Checkout the tree
	err = repo.CheckoutTree(tree, checkoutOpts)
	if err != nil {
		log.Printf("Failed to checkout tree %s", branchName)
		log.Fatal(err)
	}
	// Setting the Head to point to our branch
	repo.SetHead("refs/heads/" + branchName)

	/*
		odb, err := repo.Odb()
		if err != nil {
			log.Fatal(err)
		}
	*/

	return u1
}

func main() {
	// TODO(d34dh0r53): Implement arg parsing
	SHA = os.Args[1]
	fmt.Println(CheckOutOSA(SHA, "stable/rocky").String())
}
