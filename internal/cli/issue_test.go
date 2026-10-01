package cli

import (
	"bytes"
	"strings"
	"testing"
)

// The discovery section closes every report, so a defect report can carry a
// discovery without a second shape.
func TestIssueReportCarriesADiscoverySection(t *testing.T) {
	t.Chdir(t.TempDir())

	var out bytes.Buffer
	if err := writeReport(&out); err != nil {
		t.Fatal(err)
	}
	body := out.String()
	if !strings.Contains(body, "## 6) What I now know") {
		t.Fatalf("no discovery section:\n%s", body)
	}
	if strings.Index(body, "## 6)") < strings.Index(body, "## 5)") {
		t.Errorf("section 6 should come after section 5:\n%s", body)
	}
}
