package canary_deploy

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
	crypto_ssh "golang.org/x/crypto/ssh"
)

type GitRepoStatefileFetcher struct {
	GitRepoURL                string
	GitRepoPrivateKey         string
	GitRepoPrivateKeyPassword string
	ServiceName               string
}

func (fetcher GitRepoStatefileFetcher) Get() (Statefile, error) {
	stateFile := Statefile{}
	publicKeys, err := ssh.NewPublicKeys("git", []byte(fetcher.GitRepoPrivateKey), fetcher.GitRepoPrivateKeyPassword)
	if err != nil {
		return stateFile, fmt.Errorf("failed to extract auth info from private key: %w", err)
	}
	// required to skip the known_hosts check
	publicKeys.HostKeyCallback = crypto_ssh.InsecureIgnoreHostKey()

	// clone the repo to in-memory FS
	storer := memory.NewStorage()
	fs := memfs.New()

	_, err = git.Clone(storer, fs, &git.CloneOptions{
		Auth: publicKeys,
		URL:  fetcher.GitRepoURL,
	})
	if err != nil {
		return stateFile, fmt.Errorf("failed to clone repo: %w", err)
	}

	state, err := fs.Open(fmt.Sprintf("state-files/%s/pipeline-state.json", fetcher.ServiceName))
	if err != nil {
		return stateFile, fmt.Errorf("failed to open state file: %w", err)
	}

	byteValue, err := ioutil.ReadAll(state)
	if err != nil {
		return stateFile, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal the statefile which is in JSON format to an arbitary map.
	var data map[string]interface{}
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return stateFile, fmt.Errorf("failed to unmarshal state file to map of string and interface: %w", err)
	}
	stateFile.Data = data
	return stateFile, nil
}
