package canary_deploy

type CanaryRegionJob struct {
	StateFile    map[string]interface{}
	CanaryRegion string
	DependsOn    string
}
