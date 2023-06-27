package resource

import (
	"fmt"
	"time"

	"github.com/concourse/time-resource/canary_deploy"
	canarystatefile "github.com/concourse/time-resource/canary_state_file"
	"github.com/concourse/time-resource/lord"
	"github.com/concourse/time-resource/models"
)

type CheckCommand struct {
}

func (*CheckCommand) Run(request models.CheckRequest) ([]models.Version, error) {
	err := request.Source.Validate()
	if err != nil {
		return nil, err
	}

	previousTime := request.Version.Time
	currentTime := time.Now().UTC()

	specifiedLocation := request.Source.Location
	if specifiedLocation != nil {
		currentTime = currentTime.In((*time.Location)(specifiedLocation))
	}

	tl := lord.TimeLord{
		PreviousTime: previousTime,
		Location:     specifiedLocation,
		Start:        request.Source.Start,
		Stop:         request.Source.Stop,
		Interval:     request.Source.Interval,
		Days:         request.Source.Days,
	}

	versions := []models.Version{}

	config := canarystatefile.Config{
		GitRepoURL:                request.Source.GitRepoURL,
		GitRepoPrivateKey:         request.Source.GitRepoPrivateKey,
		GitRepoPrivateKeyPassword: request.Source.GitRepoPrivateKeyPassword,
		ServiceName:               request.Source.ServiceName,
	}
	stateFile, err := canarystatefile.GetStateFileFromGithub(config)
	if err != nil {
		return nil, fmt.Errorf("failed to get statefile: %w", err)
	}
	cd := canary_deploy.CanaryDeploy{
		StateFile:    stateFile,
		CanaryRegion: request.Source.CanaryRegion,
		DependsOn:    request.Source.DependsOn,
	}

	if !previousTime.IsZero() {
		versions = append(versions, models.Version{Time: previousTime})
	} else if request.Source.InitialVersion {
		versions = append(versions, models.Version{Time: currentTime})
		return versions, nil
	}

	if tl.Check(currentTime) && cd.Check() {
		versions = append(versions, models.Version{Time: currentTime})
	}
	return versions, nil
}
