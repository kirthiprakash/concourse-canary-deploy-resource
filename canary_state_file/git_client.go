package canarystatefile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/go-git/go-git/v5/storage/memory"
)

type Config struct {
	GitRepoURL                string
	GitRepoPrivateKey         string
	GitRepoPrivateKeyPassword string
	ServiceName               string
}

func GetStateFileFromGithub(config Config) (map[string]interface{}, error) {
	var stateFile map[string]interface{}
	publicKeys, err := ssh.NewPublicKeysFromFile("git", config.GitRepoPrivateKey, config.GitRepoPrivateKeyPassword)
	if err != nil {
		fmt.Println("publickey err", err, config.GitRepoPrivateKey)
		return stateFile, err
	}
	fmt.Println("public keys", publicKeys)
	storer := memory.NewStorage()
	fs := memfs.New()
	_, err = git.Clone(storer, fs, &git.CloneOptions{
		Auth:     publicKeys,
		URL:      config.GitRepoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return stateFile, err
	}

	state, err := fs.Open(fmt.Sprintf("state-files/%s/pipeline-state.json", config.ServiceName))
	if err != nil {
		return stateFile, err
	}

	byteValue, err := ioutil.ReadAll(state)
	if err != nil {
		return stateFile, err
	}
	json.Unmarshal(byteValue, &stateFile)
	return stateFile, nil
}
