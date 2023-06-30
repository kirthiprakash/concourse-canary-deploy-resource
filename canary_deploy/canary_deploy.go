package canary_deploy

import (
	"fmt"

	"github.com/concourse/time-resource/models"
)

type StatefileLocationType int64

const (
	GitRepo StatefileLocationType = iota
	Local
)

type Config struct {
	ReqSourcePtr *models.CanaryDeploySource
	LocationType StatefileLocationType
}

func (c Config) Validate() error {
	// Canary region check is optional.
	// Don't error if configuration related to canary region deployment is not passed.
	if c.ReqSourcePtr == nil {
		return nil
	}
	if c.ReqSourcePtr != nil && len(c.ReqSourcePtr.CanaryRegion) == 0 {
		return fmt.Errorf("field canary_deploy.canary_region is not set")
	}
	if c.ReqSourcePtr != nil && len(c.ReqSourcePtr.DependsOn) == 0 {
		return fmt.Errorf("field canary_deploy.depends_on is not set")
	}

	// GitRepoStateFile validation
	if c.LocationType == GitRepo && c.ReqSourcePtr != nil && c.ReqSourcePtr.GitRepoPtr == nil {
		return fmt.Errorf("canary_deploy.git_repo field is not set")
	}
	gitRepoPtr := c.ReqSourcePtr.GitRepoPtr
	if len(gitRepoPtr.URL) == 0 {
		return fmt.Errorf("field canary_deploy.git_repo.url is not set")
	}
	if len(gitRepoPtr.PrivateKey) == 0 {
		return fmt.Errorf("field canary_deploy.git_repo.private_key is not set")
	}
	if len(gitRepoPtr.ServiceName) == 0 {
		return fmt.Errorf("field canary_deploy.git_repo.service_name is not set")
	}
	return nil
}

func (c Config) Check() (bool, error) {
	var fetcher StateFileFetcher
	switch c.LocationType {
	case GitRepo:
		gitRepoPtr := c.ReqSourcePtr.GitRepoPtr
		fetcher = GitRepoStatefileFetcher{
			GitRepoURL:                gitRepoPtr.URL,
			GitRepoPrivateKey:         gitRepoPtr.PrivateKey,
			GitRepoPrivateKeyPassword: gitRepoPtr.PrivateKeyPassword,
			ServiceName:               gitRepoPtr.ServiceName,
		}
	case Local:
		// Sample implementation. Returns empty statefile.
		fetcher = LocalStatefileFetcher{}
	}
	if fetcher == nil {
		return false, fmt.Errorf("failed to intialise statefile fetcher. The requested statefile fetcher might not be implemented.")
	}
	statefile, err := fetcher.Get()
	if err != nil {
		return false, fmt.Errorf("failed to fetch statefile: %w", err)
	}

	forCanaryRegion := CanaryRegion{
		Name:      c.ReqSourcePtr.CanaryRegion,
		DependsOn: c.ReqSourcePtr.DependsOn,
	}
	return statefile.HasPendingDeployment(forCanaryRegion), nil
}

type CanaryRegion struct {
	Name      string
	DependsOn string
}

type CanaryRegionState struct {
	Tag    string `json:"tag"`
	Ref    string `json:"ref"`
	Paused string `json:"paused"`
}

type Statefile struct {
	Data map[string]CanaryRegionState
}

func (s Statefile) HasPendingDeployment(region CanaryRegion) bool {
	currentRegionState, ok := s.Data[region.Name]
	if !ok {
		return false
	}
	prevRegionState, ok := s.Data[region.DependsOn]
	if !ok {
		return false
	}
	if currentRegionState.Paused == "yes" {
		return false
	}
	if len(currentRegionState.Ref) > 0 && len(prevRegionState.Ref) > 0 &&
		currentRegionState.Ref != prevRegionState.Ref {
		return true
	}
	if len(currentRegionState.Tag) > 0 && len(currentRegionState.Tag) > 0 &&
		currentRegionState.Tag != prevRegionState.Tag {
		return true
	}
	return false
}

type StateFileFetcher interface {
	Get() (Statefile, error)
}
