package rules_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/smykla-labs/klaudiush/internal/rules"
)

func TestRules(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rules Suite")
}

var _ = Describe("Pattern", func() {
	Describe("DetectPatternType", func() {
		DescribeTable("should detect pattern type correctly",
			func(pattern string, expected rules.PatternType) {
				result := rules.DetectPatternType(pattern)
				Expect(result).To(Equal(expected))
			},
			// Glob patterns
			Entry("simple glob with *", "*/kong/*", rules.PatternTypeGlob),
			Entry("glob with **", "**/test/**", rules.PatternTypeGlob),
			Entry("glob with ?", "file?.txt", rules.PatternTypeGlob),
			Entry("glob with braces", "{main,master}", rules.PatternTypeGlob),
			Entry("simple path", "path/to/file", rules.PatternTypeGlob),

			// Regex patterns
			Entry("regex with ^", "^start", rules.PatternTypeRegex),
			Entry("regex with $", "end$", rules.PatternTypeRegex),
			Entry("regex with both anchors", "^exact$", rules.PatternTypeRegex),
			Entry("regex with group", "(?i)case-insensitive", rules.PatternTypeRegex),
			Entry("regex with \\d", "file\\d+", rules.PatternTypeRegex),
			Entry("regex with \\w", "\\w+", rules.PatternTypeRegex),
			Entry("regex with character class", "[a-z]+", rules.PatternTypeRegex),
			Entry("regex with alternation", "foo|bar", rules.PatternTypeRegex),
			Entry("regex with +", "a+", rules.PatternTypeRegex),
			Entry("regex with .*", "prefix.*suffix", rules.PatternTypeRegex),
			Entry("regex with .+", ".+test", rules.PatternTypeRegex),
		)
	})

	Describe("GlobPattern", func() {
		It("should match simple glob patterns", func() {
			pattern, err := rules.NewGlobPattern("**/myorg/**")
			Expect(err).NotTo(HaveOccurred())

			Expect(pattern.Match("/home/user/myorg/project")).To(BeTrue())
			Expect(pattern.Match("/home/user/other/project")).To(BeFalse())
			Expect(pattern.String()).To(Equal("**/myorg/**"))
		})

		It("should match double star patterns", func() {
			pattern, err := rules.NewGlobPattern("**/test/**")
			Expect(err).NotTo(HaveOccurred())

			Expect(pattern.Match("path/to/test/file.go")).To(BeTrue())
			Expect(pattern.Match("/test/file.go")).To(BeTrue())
			Expect(pattern.Match("path/other/file.go")).To(BeFalse())
		})

		It("should match brace expansion patterns", func() {
			pattern, err := rules.NewGlobPattern("*.{go,ts}")
			Expect(err).NotTo(HaveOccurred())

			Expect(pattern.Match("file.go")).To(BeTrue())
			Expect(pattern.Match("file.ts")).To(BeTrue())
			Expect(pattern.Match("file.js")).To(BeFalse())
		})

		It("should return error for invalid patterns", func() {
			_, err := rules.NewGlobPattern("[invalid")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("RegexPattern", func() {
		It("should match regex patterns", func() {
			pattern, err := rules.NewRegexPattern("^.*/kong/.*$")
			Expect(err).NotTo(HaveOccurred())

			Expect(pattern.Match("/home/user/kong/project")).To(BeTrue())
			Expect(pattern.Match("/home/user/other/project")).To(BeFalse())
			Expect(pattern.String()).To(Equal("^.*/kong/.*$"))
		})

		It("should match case-insensitive patterns", func() {
			pattern, err := rules.NewRegexPattern("(?i)kong")
			Expect(err).NotTo(HaveOccurred())

			Expect(pattern.Match("kong")).To(BeTrue())
			Expect(pattern.Match("Kong")).To(BeTrue())
			Expect(pattern.Match("KONG")).To(BeTrue())
		})

		It("should return error for invalid regex", func() {
			_, err := rules.NewRegexPattern("[invalid")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("CompilePattern", func() {
		It("should auto-detect and compile glob patterns", func() {
			pattern, err := rules.CompilePattern("**/myorg/**")
			Expect(err).NotTo(HaveOccurred())
			Expect(pattern.Match("/home/myorg/project")).To(BeTrue())
		})

		It("should auto-detect and compile regex patterns", func() {
			pattern, err := rules.CompilePattern("^start")
			Expect(err).NotTo(HaveOccurred())
			Expect(pattern.Match("start of string")).To(BeTrue())
			Expect(pattern.Match("middle start")).To(BeFalse())
		})
	})

	Describe("PatternCache", func() {
		var cache *rules.PatternCache

		BeforeEach(func() {
			cache = rules.NewPatternCache()
		})

		It("should cache compiled patterns", func() {
			pattern1, err1 := cache.Get("*/test/*")
			Expect(err1).NotTo(HaveOccurred())

			pattern2, err2 := cache.Get("*/test/*")
			Expect(err2).NotTo(HaveOccurred())

			// Should be the same instance.
			Expect(pattern1).To(BeIdenticalTo(pattern2))
		})

		It("should cache compilation errors", func() {
			_, err1 := cache.Get("[invalid")
			Expect(err1).To(HaveOccurred())

			_, err2 := cache.Get("[invalid")
			Expect(err2).To(HaveOccurred())
			Expect(err2).To(Equal(err1))
		})

		It("should track cache size", func() {
			Expect(cache.Size()).To(Equal(0))

			_, _ = cache.Get("pattern1")
			Expect(cache.Size()).To(Equal(1))

			_, _ = cache.Get("pattern2")
			Expect(cache.Size()).To(Equal(2))

			// Same pattern shouldn't increase size.
			_, _ = cache.Get("pattern1")
			Expect(cache.Size()).To(Equal(2))
		})

		It("should clear cache", func() {
			_, _ = cache.Get("pattern1")
			_, _ = cache.Get("pattern2")
			Expect(cache.Size()).To(Equal(2))

			cache.Clear()
			Expect(cache.Size()).To(Equal(0))
		})
	})
})
