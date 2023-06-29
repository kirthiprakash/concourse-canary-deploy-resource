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
	ReqSource    models.Source
	LocationType StatefileLocationType
}

func (c Config) Validate() error {
	if len(c.ReqSource.CanaryRegion) == 0 {
		return fmt.Errorf("field canary_region is not set")
	}
	if len(c.ReqSource.DependsOn) == 0 {
		return fmt.Errorf("field depends_on is not set")
	}
	// GitRepoStateFile validation
	if len(c.ReqSource.GitRepoURL) == 0 {
		return fmt.Errorf("field git_repo_url is not set")
	}
	if len(c.ReqSource.GitRepoPrivateKey) == 0 {
		return fmt.Errorf("field git_repo_private_key is not set")
	}
	if len(c.ReqSource.ServiceName) == 0 {
		return fmt.Errorf("field service_name is not set")
	}
	return nil
}

func (c Config) Check() (bool, error) {
	var fetcher StateFileFetcher
	switch c.LocationType {
	case GitRepo:
		fetcher = GitRepoStatefileFetcher{
			GitRepoURL:                c.ReqSource.GitRepoURL,
			GitRepoPrivateKey:         c.ReqSource.GitRepoPrivateKey,
			GitRepoPrivateKeyPassword: c.ReqSource.GitRepoPrivateKeyPassword,
			ServiceName:               c.ReqSource.ServiceName,
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
		Name:      c.ReqSource.CanaryRegion,
		DependsOn: c.ReqSource.DependsOn,
	}
	return statefile.HasPendingDeployment(forCanaryRegion), nil

}

type CanaryRegion struct {
	Name      string
	DependsOn string
}

//
type Statefile struct {
	Data map[string]interface{}
}

func (s Statefile) HasPendingDeployment(region CanaryRegion) bool {
	fmt.Println(s.Data)
	currentRegionStateIn, ok := s.Data[region.Name]
	if !ok {
		return false
	}
	prevRegionStateIn, ok := s.Data[region.DependsOn]
	if !ok {
		return false
	}
	currentRegionState, ok := currentRegionStateIn.(map[string]interface{})
	if !ok {
		return false
	}
	prevRegionState, ok := prevRegionStateIn.(map[string]interface{})
	if !ok {
		return false
	}
	currentRegionTag, ok := currentRegionState["tag"].(string)
	if !ok {
		return false
	}
	prevRegionTag, ok := prevRegionState["tag"].(string)
	if !ok {
		return false
	}
	if len(currentRegionTag) > 0 && len(prevRegionTag) > 0 && currentRegionTag != prevRegionTag {
		return true
	}
	return false
}

type StateFileFetcher interface {
	Get() (Statefile, error)
}
