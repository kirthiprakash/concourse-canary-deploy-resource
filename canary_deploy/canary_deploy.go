package canary_deploy

type CanaryDeploy struct {
	StateFile    map[string]interface{}
	CanaryRegion string
	DependsOn    string
}

func (cd CanaryDeploy) Check() bool {
	currentRegionStateIn, ok := cd.StateFile[cd.CanaryRegion]
	if !ok {
		return false
	}
	prevRegionStateIn, ok := cd.StateFile[cd.DependsOn]
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
