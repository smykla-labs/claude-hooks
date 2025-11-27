package config_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/klaudiush/pkg/config"
)

func TestConfigRules(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Rules Suite")
}

var _ = Describe("RulesConfig", func() {
	Describe("IsEnabled", func() {
		It("should return true when Enabled is nil", func() {
			cfg := &config.RulesConfig{}
			Expect(cfg.IsEnabled()).To(BeTrue())
		})

		It("should return true when Enabled is true", func() {
			enabled := true
			cfg := &config.RulesConfig{Enabled: &enabled}
			Expect(cfg.IsEnabled()).To(BeTrue())
		})

		It("should return false when Enabled is false", func() {
			enabled := false
			cfg := &config.RulesConfig{Enabled: &enabled}
			Expect(cfg.IsEnabled()).To(BeFalse())
		})

		It("should return true for nil RulesConfig", func() {
			var cfg *config.RulesConfig
			Expect(cfg.IsEnabled()).To(BeTrue())
		})
	})

	Describe("ShouldStopOnFirstMatch", func() {
		It("should return true when StopOnFirstMatch is nil", func() {
			cfg := &config.RulesConfig{}
			Expect(cfg.ShouldStopOnFirstMatch()).To(BeTrue())
		})

		It("should return true when StopOnFirstMatch is true", func() {
			stop := true
			cfg := &config.RulesConfig{StopOnFirstMatch: &stop}
			Expect(cfg.ShouldStopOnFirstMatch()).To(BeTrue())
		})

		It("should return false when StopOnFirstMatch is false", func() {
			stop := false
			cfg := &config.RulesConfig{StopOnFirstMatch: &stop}
			Expect(cfg.ShouldStopOnFirstMatch()).To(BeFalse())
		})

		It("should return true for nil RulesConfig", func() {
			var cfg *config.RulesConfig
			Expect(cfg.ShouldStopOnFirstMatch()).To(BeTrue())
		})
	})
})

var _ = Describe("RuleConfig", func() {
	Describe("IsRuleEnabled", func() {
		It("should return true when Enabled is nil", func() {
			cfg := config.RuleConfig{}
			Expect(cfg.IsRuleEnabled()).To(BeTrue())
		})

		It("should return true when Enabled is true", func() {
			enabled := true
			cfg := config.RuleConfig{Enabled: &enabled}
			Expect(cfg.IsRuleEnabled()).To(BeTrue())
		})

		It("should return false when Enabled is false", func() {
			enabled := false
			cfg := config.RuleConfig{Enabled: &enabled}
			Expect(cfg.IsRuleEnabled()).To(BeFalse())
		})
	})
})

var _ = Describe("RuleActionConfig", func() {
	Describe("GetActionType", func() {
		It("should return 'block' when Type is empty", func() {
			cfg := &config.RuleActionConfig{}
			Expect(cfg.GetActionType()).To(Equal("block"))
		})

		It("should return 'block' for nil config", func() {
			var cfg *config.RuleActionConfig
			Expect(cfg.GetActionType()).To(Equal("block"))
		})

		It("should return the configured type", func() {
			cfg := &config.RuleActionConfig{Type: "warn"}
			Expect(cfg.GetActionType()).To(Equal("warn"))
		})

		It("should return 'allow' when configured", func() {
			cfg := &config.RuleActionConfig{Type: "allow"}
			Expect(cfg.GetActionType()).To(Equal("allow"))
		})
	})
})
