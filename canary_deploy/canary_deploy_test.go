package canary_deploy_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/concourse/time-resource/canary_deploy"
	"github.com/concourse/time-resource/models"
)

var _ = Describe("Config", func() {
	Describe("Validate", func() {
		var config canary_deploy.Config
		var emptyDependsOn,
			emptyCanaryRegion,
			emptyGitRepo,
			emptyGitRepoUrl,
			emptyGitRepoPrivateKey,
			emptyGitRepoServiceName,
			emptyGitRepoSHHPassword models.CanaryDeploySource

		BeforeEach(func() {
			config = canary_deploy.Config{}
			emptyDependsOn = models.CanaryDeploySource{
				CanaryRegion: "something",
				DependsOn:    "",
			}
			emptyCanaryRegion = models.CanaryDeploySource{
				CanaryRegion: "",
				DependsOn:    "something",
			}
			emptyGitRepo = models.CanaryDeploySource{
				CanaryRegion: "Something",
				DependsOn:    "something",
				GitRepoPtr:   nil,
			}
			emptyGitRepoUrl = models.CanaryDeploySource{
				CanaryRegion: "Something",
				DependsOn:    "something",
				GitRepoPtr: &models.GitRepoSource{
					URL: "",
				},
			}
			emptyGitRepoPrivateKey = models.CanaryDeploySource{
				CanaryRegion: "Something",
				DependsOn:    "something",
				GitRepoPtr: &models.GitRepoSource{
					URL:        "something",
					PrivateKey: "",
				},
			}
			emptyGitRepoServiceName = models.CanaryDeploySource{
				CanaryRegion: "Something",
				DependsOn:    "something",
				GitRepoPtr: &models.GitRepoSource{
					URL:         "something",
					PrivateKey:  "something",
					ServiceName: "",
				},
			}
			emptyGitRepoSHHPassword = models.CanaryDeploySource{
				CanaryRegion: "Something",
				DependsOn:    "something",
				GitRepoPtr: &models.GitRepoSource{
					URL:                "something",
					PrivateKey:         "something",
					ServiceName:        "something",
					PrivateKeyPassword: "",
				},
			}
		})
		It("errors on empty canary_region", func() {
			config.ReqSourcePtr = &emptyCanaryRegion
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("canary_deploy.canary_region"))
		})
		It("errors on empty depends_on", func() {
			config.ReqSourcePtr = &emptyDependsOn
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("canary_deploy.depends_on"))
		})
		It("doesn't error if canary region check is disabled", func() {
			config.ReqSourcePtr = nil
			err := config.Validate()
			Expect(err).To(BeNil())
		})
		It("errors on empty git_repo field", func() {
			config.ReqSourcePtr = &emptyGitRepo
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("canary_deploy.git_repo"))
		})
		It("errors on empty git_repo URL field", func() {
			config.ReqSourcePtr = &emptyGitRepoUrl
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("canary_deploy.git_repo.url"))
		})
		It("errors on empty git_repo private key field", func() {
			config.ReqSourcePtr = &emptyGitRepoPrivateKey
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("canary_deploy.git_repo.private_key"))
		})
		It("errors on empty git_repo service name field", func() {
			config.ReqSourcePtr = &emptyGitRepoServiceName
			err := config.Validate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("canary_deploy.git_repo.service_name"))
		})
		It("doesn't error on empty ssh passphase", func() {
			config.ReqSourcePtr = &emptyGitRepoSHHPassword
			err := config.Validate()
			Expect(err).To(BeNil())
		})
	})
})

var _ = Describe("CanaryDeploy", func() {
	Describe("Statefile Check", func() {
		var statefile canary_deploy.Statefile
		var currentRegion canary_deploy.CanaryRegion
		var tagVer1, tagVer2, refA, refB, deployPaused canary_deploy.CanaryRegionState
		BeforeEach(func() {
			statefile = canary_deploy.Statefile{}
			currentRegion = canary_deploy.CanaryRegion{
				Name:      "ap-southeast-2",
				DependsOn: "prep",
			}

			tagVer1 = canary_deploy.CanaryRegionState{
				Tag:    "ver1",
				Ref:    "random",
				Paused: "no",
			}
			tagVer2 = canary_deploy.CanaryRegionState{
				Tag:    "ver2",
				Ref:    "random",
				Paused: "no",
			}
			refA = canary_deploy.CanaryRegionState{
				Tag:    "random",
				Ref:    "A",
				Paused: "no",
			}
			refB = canary_deploy.CanaryRegionState{
				Tag:    "random",
				Ref:    "B",
				Paused: "no",
			}
			deployPaused = canary_deploy.CanaryRegionState{
				Tag:    "random",
				Ref:    "random",
				Paused: "yes",
			}

		})

		It("returns false on empty statefile", func() {

			sf := canary_deploy.Statefile{
				Data: make(map[string]canary_deploy.CanaryRegionState),
			}
			ret := sf.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns false on invalid region", func() {

			sf := canary_deploy.Statefile{
				Data: map[string]canary_deploy.CanaryRegionState{
					"invalid": tagVer1,
				},
			}
			ret := sf.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns false on region being paused - invalid depends on", func() {

			sf := canary_deploy.Statefile{
				Data: map[string]canary_deploy.CanaryRegionState{
					currentRegion.Name: deployPaused,
				},
			}
			ret := sf.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns false on region being paused - no version change", func() {

			sf := canary_deploy.Statefile{
				Data: map[string]canary_deploy.CanaryRegionState{
					currentRegion.Name:      deployPaused,
					currentRegion.DependsOn: deployPaused,
				},
			}
			ret := sf.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns false on region being paused - version change", func() {

			sf := canary_deploy.Statefile{
				Data: map[string]canary_deploy.CanaryRegionState{
					currentRegion.Name:      deployPaused,
					currentRegion.DependsOn: tagVer1,
				},
			}
			ret := sf.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns false on no verion difference - ref", func() {
			statefile.Data = map[string]canary_deploy.CanaryRegionState{
				currentRegion.Name:      refA,
				currentRegion.DependsOn: refA,
			}
			ret := statefile.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns false on no verion difference - tag", func() {
			statefile.Data = map[string]canary_deploy.CanaryRegionState{
				currentRegion.Name:      tagVer1,
				currentRegion.DependsOn: tagVer1,
			}
			ret := statefile.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeFalse())
		})
		It("returns true on verion difference - ref", func() {
			statefile.Data = map[string]canary_deploy.CanaryRegionState{
				currentRegion.Name:      refA,
				currentRegion.DependsOn: refB,
			}
			ret := statefile.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeTrue())
		})
		It("returns true on verion difference - tag", func() {
			statefile.Data = map[string]canary_deploy.CanaryRegionState{
				currentRegion.Name:      tagVer1,
				currentRegion.DependsOn: tagVer2,
			}
			ret := statefile.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeTrue())
		})
		It("returns true on verion difference - tag & ref", func() {
			statefile.Data = map[string]canary_deploy.CanaryRegionState{
				currentRegion.Name:      tagVer1,
				currentRegion.DependsOn: refA,
			}
			ret := statefile.HasPendingDeployment(currentRegion)
			Expect(ret).To(BeTrue())
		})
	})
})
