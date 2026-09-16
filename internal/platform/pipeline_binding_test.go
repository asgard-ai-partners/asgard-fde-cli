package platform

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// bindingTags are the two fields that decide which repository a pipeline reads,
// by the names the platform gives them.
var bindingTags = []string{"connection_id", "repository_id"}

// TestNoPipelineUpdateTakesTheBinding refuses a pipeline update that carries
// the repository binding.
//
// **The absence of a `pipeline update` is not what stops this**, because that
// absence is temporary: the platform has a PATCH, and whoever wraps it will
// write the input type by reading CreatePipelineInput and keeping the fields
// that look like settings. These two are not settings. Moving a pipeline to
// another repository is its own pair of calls on the platform -
// `preflight-repository`, then `change-repository` - and the preflight is the
// only part of it that is a decision: it compares the candidate's declaration
// with this pipeline release by release, while the releases themselves keep
// their platform Projects, namespaces, variables and deploy identities either
// way. A field riding along on an update is that pair with the comparison
// deleted.
//
// A wrapper for the pair is a different thing and is not refused here: it takes
// the two calls and the comparison, and an input named for what it does -
// ChangePipelineRepositoryInput - is not an update and does not match below.
func TestNoPipelineUpdateTakesTheBinding(t *testing.T) {
	// **A checker that finds nothing to check passes everything.** If the
	// platform renames these fields, everything below goes on passing while
	// reading for words that are no longer written anywhere - so the words are
	// held against the one type that must carry them.
	create := reflect.TypeOf(CreatePipelineInput{})
	for _, tag := range bindingTags {
		if !hasJSONTag(create, tag) {
			t.Fatalf("CreatePipelineInput no longer has a %q field, so this test is reading for a name that has moved; "+
				"find what the platform calls it now and change bindingTags", tag)
		}
	}

	for _, s := range packageStructs(t) {
		if !updatesAPipeline(s.name) {
			continue
		}
		for _, f := range s.node.Fields.List {
			tag := jsonTag(f)
			for _, bad := range bindingTags {
				if tag != bad {
					continue
				}
				t.Errorf("%s carries %s, which re-binds the pipeline to another repository.\n"+
					"The platform does that through preflight-repository and change-repository, in that order, "+
					"and the preflight's comparison is what makes it safe; an update field is that pair with the "+
					"comparison dropped. Wrap the pair instead, under a name that says so.", s.name, tag)
			}
		}
	}
}

// updatesAPipeline reads a type name as a claim about what it does. **The name
// is the check**, because the mistake is one of altitude rather than of
// spelling: somebody puts a field where it does not belong, and what tells them
// so is being made to say out loud that an update is what they are writing.
func updatesAPipeline(name string) bool {
	if !strings.Contains(name, "Pipeline") {
		return false
	}
	for _, verb := range []string{"Update", "Patch", "Modify", "Edit"} {
		if strings.Contains(name, verb) {
			return true
		}
	}
	return false
}

type namedStruct struct {
	name string
	node *ast.StructType
}

// packageStructs returns every struct declared in this package, from source
// rather than from reflection: a type nothing constructs is invisible to
// reflect, and a type nothing constructs yet is exactly the one this is for.
func packageStructs(t *testing.T) []namedStruct {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}
	var out []namedStruct
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, filepath.Join(".", name), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(file, func(n ast.Node) bool {
			ts, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok || st.Fields == nil {
				return true
			}
			out = append(out, namedStruct{name: ts.Name.Name, node: st})
			return true
		})
	}
	if len(out) == 0 {
		t.Fatal("no struct found in this package, so this test read nothing")
	}
	return out
}

// jsonTag is the name one field is serialised under, without its options.
func jsonTag(f *ast.Field) string {
	if f.Tag == nil {
		return ""
	}
	raw, err := strconv.Unquote(f.Tag.Value)
	if err != nil {
		return ""
	}
	name, _, _ := strings.Cut(reflect.StructTag(raw).Get("json"), ",")
	return name
}

func hasJSONTag(t reflect.Type, want string) bool {
	for i := 0; i < t.NumField(); i++ {
		name, _, _ := strings.Cut(t.Field(i).Tag.Get("json"), ",")
		if name == want {
			return true
		}
	}
	return false
}
