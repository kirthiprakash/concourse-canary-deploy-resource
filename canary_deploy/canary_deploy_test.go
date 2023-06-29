package canary_deploy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/time-resource/canary_deploy"
)

var _ = Describe("CanaryDeploy", func() {
	Describe("Check", func() {
		It("returns false on empty statefile", func() {
			forRegion := canary_deploy.CanaryRegion{
				Name:      "ap-southeast-2",
				DependsOn: "prep",
			}
			sf := canary_deploy.Statefile{
				Data: make(map[string]interface{}),
			}
			ret := sf.HasPendingDeployment(forRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns true on version difference", func() {
			forRegion := canary_deploy.CanaryRegion{
				Name:      "ap-southeast-2",
				DependsOn: "prep",
			}
			sf := canary_deploy.Statefile{
				Data: map[string]interface{}{
					"ap-southeast-2": map[string]interface{}{
						"tag": "0.0.1",
					},
					"prep": map[string]interface{}{
						"tag": "0.0.2",
					},
				},
			}
			ret := sf.HasPendingDeployment(forRegion)
			Expect(ret).To(BeTrue())
		})
	})
})
