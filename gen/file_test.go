package gen

import (
	"strings"
	"testing"

	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/exp"
	"github.com/mb0/xelf/typ"
)

func TestWriteGoFile(t *testing.T) {
	align := typ.Flag("foo.Align")
	align.Consts = []cor.Const{{"A", 1}, {"B", 2}, {"C", 3}}
	kind := typ.Enum("foo.Kind")
	kind.Consts = []cor.Const{{"A", 1}, {"B", 2}, {"C", 3}}
	tests := []struct {
		els  []exp.El
		want string
	}{
		{nil, "package foo\n"},
		{[]exp.El{align},
			"package foo\n\ntype Align uint64\n\n" +
				"const (\n" +
				"\tAlignA Align = 1 << iota\n" +
				"\tAlignB\n" +
				"\tAlignC = AlignA | AlignB\n" +
				")\n",
		},
		{[]exp.El{kind},
			"package foo\n\ntype Kind string\n\n" +
				"const (\n" +
				"\tKindA Kind = \"a\"\n" +
				"\tKindB Kind = \"b\"\n" +
				"\tKindC Kind = \"c\"\n" +
				")\n",
		},
		{[]exp.El{typ.Rec("foo.Node", []typ.Field{
			{Name: "Name?", Type: typ.Str},
		})}, "package foo\n\ntype Node struct {\n" +
			"\tName string `json:\"name,omitempty\"`\n" + "}\n",
		},
		{[]exp.El{typ.Rec("foo.Node", []typ.Field{
			{Name: "Start", Type: typ.Time},
		})}, "package foo\n\nimport (\n\t\"time\"\n)\n\ntype Node struct {\n" +
			"\tStart time.Time `json:\"start\"`\n" + "}\n",
		},
		{[]exp.El{typ.Rec("foo.Node", []typ.Field{
			{Name: "Kind", Type: typ.Enum("bar.Kind")},
		})}, "package foo\n\nimport (\n\t\"path/to/bar\"\n)\n\ntype Node struct {\n" +
			"\tKind bar.Kind `json:\"kind\"`\n" + "}\n",
		},
		{[]exp.El{typ.Rec("foo.Node", []typ.Field{
			{Name: "Kind", Type: typ.Enum("foo.Kind")},
		})}, "package foo\n\ntype Node struct {\n" +
			"\tKind Kind `json:\"kind\"`\n" + "}\n",
		},
	}
	pkgs := map[string]string{
		"cor": "github.com/mb0/xelf/cor",
		"foo": "path/to/foo",
		"bar": "path/to/bar",
	}
	for _, test := range tests {
		var b strings.Builder
		c := &Ctx{B: &b, Pkg: "path/to/foo", Pkgs: pkgs}
		err := WriteGoFile(c, test.els)
		if err != nil {
			t.Errorf("write %+v error: %v", test.els, err)
			continue
		}
		if got := b.String(); got != test.want {
			t.Errorf("for %+v want %s got %s", test.els, test.want, got)
		}
	}
}
