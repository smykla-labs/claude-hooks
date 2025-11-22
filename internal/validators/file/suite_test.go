package file_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFileValidators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "File Validators Suite")
}
