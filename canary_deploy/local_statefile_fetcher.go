package canary_deploy

type LocalStatefileFetcher struct {
}

func (fetcher LocalStatefileFetcher) Get() (Statefile, error) {
	return Statefile{
		Data: make(map[string]interface{}),
	}, nil
}
