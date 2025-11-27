package rules_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/klaudiush/internal/rules"
	"github.com/smykla-labs/klaudiush/pkg/hook"
)

var _ = Describe("Matcher", func() {
	Describe("RepoPatternMatcher", func() {
		It("should match repo root with glob pattern", func() {
			matcher, err := rules.NewRepoPatternMatcher("**/myorg/**")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				GitContext: &rules.GitContext{
					RepoRoot: "/home/user/myorg/project",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
			Expect(matcher.Name()).To(ContainSubstring("repo_pattern"))
		})

		It("should not match when GitContext is nil", func() {
			matcher, err := rules.NewRepoPatternMatcher("**/myorg/**")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{}
			Expect(matcher.Match(ctx)).To(BeFalse())
		})

		It("should match with regex pattern", func() {
			matcher, err := rules.NewRepoPatternMatcher("(?i).*/myorg/.*")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				GitContext: &rules.GitContext{
					RepoRoot: "/home/user/MyOrg/project",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})
	})

	Describe("RemoteMatcher", func() {
		It("should match exact remote name", func() {
			matcher := rules.NewRemoteMatcher("origin")

			ctx := &rules.MatchContext{
				GitContext: &rules.GitContext{
					Remote: "origin",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
			Expect(matcher.Name()).To(Equal("remote:origin"))
		})

		It("should not match different remote", func() {
			matcher := rules.NewRemoteMatcher("origin")

			ctx := &rules.MatchContext{
				GitContext: &rules.GitContext{
					Remote: "upstream",
				},
			}
			Expect(matcher.Match(ctx)).To(BeFalse())
		})

		It("should not match when GitContext is nil", func() {
			matcher := rules.NewRemoteMatcher("origin")

			ctx := &rules.MatchContext{}
			Expect(matcher.Match(ctx)).To(BeFalse())
		})
	})

	Describe("BranchPatternMatcher", func() {
		It("should match branch with glob pattern", func() {
			matcher, err := rules.NewBranchPatternMatcher("feature/*")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				GitContext: &rules.GitContext{
					Branch: "feature/new-feature",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})

		It("should match branch with regex pattern", func() {
			matcher, err := rules.NewBranchPatternMatcher("^release-\\d+\\.\\d+$")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				GitContext: &rules.GitContext{
					Branch: "release-1.2",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())

			ctx.GitContext.Branch = "release-test"
			Expect(matcher.Match(ctx)).To(BeFalse())
		})
	})

	Describe("FilePatternMatcher", func() {
		It("should match file path from FileContext", func() {
			matcher, err := rules.NewFilePatternMatcher("**/test/**")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				FileContext: &rules.FileContext{
					Path: "src/test/file.go",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})

		It("should fall back to HookContext file path", func() {
			matcher, err := rules.NewFilePatternMatcher("*.go")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				HookContext: &hook.Context{
					ToolInput: hook.ToolInput{
						FilePath: "main.go",
					},
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})
	})

	Describe("ContentPatternMatcher", func() {
		It("should match content with regex", func() {
			matcher, err := rules.NewContentPatternMatcher("(?i)password")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				FileContext: &rules.FileContext{
					Content: "const PASSWORD = 'secret'",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})

		It("should fall back to HookContext content", func() {
			matcher, err := rules.NewContentPatternMatcher("func main")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				HookContext: &hook.Context{
					ToolInput: hook.ToolInput{
						Content: "package main\n\nfunc main() {}",
					},
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})
	})

	Describe("CommandPatternMatcher", func() {
		It("should match command with glob pattern", func() {
			matcher, err := rules.NewCommandPatternMatcher("git push*")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				Command: "git push origin main",
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})

		It("should fall back to HookContext command", func() {
			matcher, err := rules.NewCommandPatternMatcher("git*")
			Expect(err).NotTo(HaveOccurred())

			ctx := &rules.MatchContext{
				HookContext: &hook.Context{
					ToolInput: hook.ToolInput{
						Command: "git status",
					},
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})
	})

	Describe("ValidatorTypeMatcher", func() {
		It("should match exact validator type", func() {
			matcher := rules.NewValidatorTypeMatcher(rules.ValidatorGitPush)

			ctx := &rules.MatchContext{
				ValidatorType: rules.ValidatorGitPush,
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})

		It("should match wildcard all", func() {
			matcher := rules.NewValidatorTypeMatcher(rules.ValidatorAll)

			ctx := &rules.MatchContext{
				ValidatorType: rules.ValidatorGitPush,
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})

		It("should match category wildcard", func() {
			matcher := rules.NewValidatorTypeMatcher(rules.ValidatorGitAll)

			ctx := &rules.MatchContext{
				ValidatorType: rules.ValidatorGitPush,
			}
			Expect(matcher.Match(ctx)).To(BeTrue())

			ctx.ValidatorType = rules.ValidatorGitCommit
			Expect(matcher.Match(ctx)).To(BeTrue())

			ctx.ValidatorType = rules.ValidatorFileMarkdown
			Expect(matcher.Match(ctx)).To(BeFalse())
		})

		It("should not match when ValidatorType is empty", func() {
			matcher := rules.NewValidatorTypeMatcher(rules.ValidatorGitPush)

			ctx := &rules.MatchContext{}
			Expect(matcher.Match(ctx)).To(BeFalse())
		})
	})

	Describe("ToolTypeMatcher", func() {
		It("should match tool type case-insensitively", func() {
			matcher := rules.NewToolTypeMatcher("Bash")

			ctx := &rules.MatchContext{
				HookContext: &hook.Context{
					ToolName: hook.ToolTypeBash,
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})
	})

	Describe("EventTypeMatcher", func() {
		It("should match event type case-insensitively", func() {
			matcher := rules.NewEventTypeMatcher("PreToolUse")

			ctx := &rules.MatchContext{
				HookContext: &hook.Context{
					EventType: hook.EventTypePreToolUse,
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())
		})
	})

	Describe("CompositeMatcher", func() {
		Describe("AND", func() {
			It("should match when all conditions match", func() {
				matcher := rules.NewAndMatcher(
					rules.NewRemoteMatcher("origin"),
					rules.NewValidatorTypeMatcher(rules.ValidatorGitPush),
				)

				ctx := &rules.MatchContext{
					ValidatorType: rules.ValidatorGitPush,
					GitContext: &rules.GitContext{
						Remote: "origin",
					},
				}
				Expect(matcher.Match(ctx)).To(BeTrue())
				Expect(matcher.Name()).To(Equal("AND"))
			})

			It("should not match when any condition fails", func() {
				matcher := rules.NewAndMatcher(
					rules.NewRemoteMatcher("origin"),
					rules.NewValidatorTypeMatcher(rules.ValidatorGitPush),
				)

				ctx := &rules.MatchContext{
					ValidatorType: rules.ValidatorGitPush,
					GitContext: &rules.GitContext{
						Remote: "upstream",
					},
				}
				Expect(matcher.Match(ctx)).To(BeFalse())
			})

			It("should match with empty matchers", func() {
				matcher := rules.NewAndMatcher()
				ctx := &rules.MatchContext{}
				Expect(matcher.Match(ctx)).To(BeTrue())
			})
		})

		Describe("OR", func() {
			It("should match when any condition matches", func() {
				matcher := rules.NewOrMatcher(
					rules.NewRemoteMatcher("origin"),
					rules.NewRemoteMatcher("upstream"),
				)

				ctx := &rules.MatchContext{
					GitContext: &rules.GitContext{
						Remote: "upstream",
					},
				}
				Expect(matcher.Match(ctx)).To(BeTrue())
				Expect(matcher.Name()).To(Equal("OR"))
			})

			It("should not match when no conditions match", func() {
				matcher := rules.NewOrMatcher(
					rules.NewRemoteMatcher("origin"),
					rules.NewRemoteMatcher("upstream"),
				)

				ctx := &rules.MatchContext{
					GitContext: &rules.GitContext{
						Remote: "other",
					},
				}
				Expect(matcher.Match(ctx)).To(BeFalse())
			})
		})

		Describe("NOT", func() {
			It("should invert the result", func() {
				matcher := rules.NewNotMatcher(
					rules.NewRemoteMatcher("origin"),
				)

				ctx := &rules.MatchContext{
					GitContext: &rules.GitContext{
						Remote: "origin",
					},
				}
				Expect(matcher.Match(ctx)).To(BeFalse())

				ctx.GitContext.Remote = "upstream"
				Expect(matcher.Match(ctx)).To(BeTrue())
				Expect(matcher.Name()).To(Equal("NOT"))
			})
		})
	})

	Describe("BuildMatcher", func() {
		It("should build composite matcher from RuleMatch", func() {
			match := &rules.RuleMatch{
				ValidatorType: rules.ValidatorGitPush,
				RepoPattern:   "**/myorg/**",
				Remote:        "origin",
			}

			matcher, err := rules.BuildMatcher(match)
			Expect(err).NotTo(HaveOccurred())
			Expect(matcher).NotTo(BeNil())

			ctx := &rules.MatchContext{
				ValidatorType: rules.ValidatorGitPush,
				GitContext: &rules.GitContext{
					RepoRoot: "/home/user/myorg/project",
					Remote:   "origin",
				},
			}
			Expect(matcher.Match(ctx)).To(BeTrue())

			// Should not match with different remote.
			ctx.GitContext.Remote = "upstream"
			Expect(matcher.Match(ctx)).To(BeFalse())
		})

		It("should return nil for nil RuleMatch", func() {
			matcher, err := rules.BuildMatcher(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(matcher).To(BeNil())
		})

		It("should return nil for empty RuleMatch", func() {
			matcher, err := rules.BuildMatcher(&rules.RuleMatch{})
			Expect(err).NotTo(HaveOccurred())
			Expect(matcher).To(BeNil())
		})

		It("should return single matcher for single condition", func() {
			match := &rules.RuleMatch{
				Remote: "origin",
			}

			matcher, err := rules.BuildMatcher(match)
			Expect(err).NotTo(HaveOccurred())
			Expect(matcher).NotTo(BeNil())

			// Should be a RemoteMatcher, not a CompositeMatcher.
			Expect(matcher.Name()).To(Equal("remote:origin"))
		})

		It("should return error for invalid pattern", func() {
			match := &rules.RuleMatch{
				RepoPattern: "[invalid",
			}

			_, err := rules.BuildMatcher(match)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("AlwaysMatcher", func() {
		It("should always match", func() {
			matcher := &rules.AlwaysMatcher{}
			Expect(matcher.Match(&rules.MatchContext{})).To(BeTrue())
			Expect(matcher.Match(nil)).To(BeTrue())
			Expect(matcher.Name()).To(Equal("always"))
		})
	})

	Describe("NeverMatcher", func() {
		It("should never match", func() {
			matcher := &rules.NeverMatcher{}
			Expect(matcher.Match(&rules.MatchContext{})).To(BeFalse())
			Expect(matcher.Match(nil)).To(BeFalse())
			Expect(matcher.Name()).To(Equal("never"))
		})
	})
})
