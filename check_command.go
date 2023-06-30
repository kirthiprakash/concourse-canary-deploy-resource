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

	// Validating inputs for canary deployment if provided
	// Q: Why validate here? Why not just before calling Check()?
	// A: Fail fast. The Canary Deploy check is called only when the
	//    time criteria is satisfied.
	var canaryDeploy canary_deploy.Config
	if request.Source.CanaryDeployPtr != nil {
		canaryDeploy = canary_deploy.Config{
			ReqSourcePtr: request.Source.CanaryDeployPtr,
			LocationType: canary_deploy.GitRepo,
		}

		err = canaryDeploy.Validate()
		if err != nil {
			return nil, err
		}
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
		hasPendingDeployment := false
		if request.Source.CanaryDeployPtr != nil {
			hasPendingDeployment, err = canaryDeploy.Check()
			if err != nil {
				return nil, fmt.Errorf("failed to check canary deploy statefile. err: %q", err)
			}
		}

		if request.Source.CanaryDeployPtr == nil || hasPendingDeployment {
			versions = append(versions, models.Version{Time: currentTime})
		}
	}
	return versions, nil
}
