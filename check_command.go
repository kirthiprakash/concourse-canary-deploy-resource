package resource

import (
	"fmt"
	"time"

	"github.com/concourse/time-resource/canary_deploy"
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

	canaryDeploy := canary_deploy.Config{
		ReqSource:    request.Source,
		LocationType: canary_deploy.GitRepo,
	}
	// validating inputs for canary deployment
	err = canaryDeploy.Validate()
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

	if !previousTime.IsZero() {
		versions = append(versions, models.Version{Time: previousTime})
	} else if request.Source.InitialVersion {
		versions = append(versions, models.Version{Time: currentTime})
		return versions, nil
	}

	if tl.Check(currentTime) {

		hasPendingDeployment, err := canaryDeploy.Check()
		if err != nil {
			return nil, fmt.Errorf("failed to check canary deploy statefile. err: %q", err)
		}
		// Append new version only if both time and canary deploy criteria are true.
		if hasPendingDeployment {
			versions = append(versions, models.Version{Time: currentTime})
		}
	}
	return versions, nil
}
