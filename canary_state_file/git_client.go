package canarystatefile

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

type Config struct {
	GitRepoURL                string
	GitRepoPrivateKey         string
	GitRepoPrivateKeyPassword string
	ServiceName               string
}

func GetStateFileFromGithub(config Config) (map[string]interface{}, error) {
	var stateFile map[string]interface{}
	publicKeys, err := ssh.NewPublicKeys("git", []byte(config.GitRepoPrivateKey), config.GitRepoPrivateKeyPassword)
	if err != nil {
		return stateFile, fmt.Errorf("failed to extract auth info from private key: %w", err)
	}
	publicKeys.HostKeyCallback = crypto_ssh.InsecureIgnoreHostKey()
	storer := memory.NewStorage()
	fs := memfs.New()
	_, err = git.Clone(storer, fs, &git.CloneOptions{
		Auth: publicKeys,
		URL:  config.GitRepoURL,
	})
	if err != nil {
		return stateFile, fmt.Errorf("failed to clone repo: %w", err)
	}

	state, err := fs.Open(fmt.Sprintf("state-files/%s/pipeline-state.json", config.ServiceName))
	if err != nil {
		return stateFile, fmt.Errorf("failed to open state file: %w", err)
	}

	byteValue, err := ioutil.ReadAll(state)
	if err != nil {
		return stateFile, fmt.Errorf("failed to read state file: %w", err)
	}
	err = json.Unmarshal(byteValue, &stateFile)
	if err != nil {
		return stateFile, fmt.Errorf("failed to unmarshal state file to map of string and interface: %w", err)
	}
	return stateFile, nil
}
