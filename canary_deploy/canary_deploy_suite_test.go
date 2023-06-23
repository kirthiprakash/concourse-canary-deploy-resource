package canary_deploy_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCanaryDeploy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CanaryDeploy Suite")
}
