package canary_deploy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/time-resource/canary_deploy"
)

var _ = Describe("CanaryDeploy", func() {
	Describe("Check", func() {
		It("returns false on empty statefile", func() {
			cd := canary_deploy.CanaryDeploy{
				CanaryRegion: "ap-southeast-2",
				DependsOn:    "prep",
			}
			ret := cd.Check()
			Expect(ret).To(BeFalse())
		})
		It("returns true on version difference", func() {
			cd := canary_deploy.CanaryDeploy{
				CanaryRegion: "ap-southeast-2",
				DependsOn:    "prep",
				StateFile: map[string]interface{}{
					"ap-southeast-2": map[string]string{
						"tag": "0.0.1",
					},
					"prep": map[string]string{
						"tag": "0.0.2",
					},
				},
			}
			ret := cd.Check()
			Expect(ret).To(BeTrue())
		})
	})
})
