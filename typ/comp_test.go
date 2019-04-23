package typ

import "testing"

func TestComp(t *testing.T) {
	tests := []struct {
		want     Cmp
		src, dst string
	}{
		{CmpSame, "int", "int"},
		{CmpSame, "list|int", "list|int"},
		{CmpSame, "any", "any"},
		{CmpInfer, "any", "@"},
		{CmpInfer, "any", "@1"},
		{CmpCheckRef, "@ref", "@ref"},
		{CmpInfer, "@", "@"},
		{CmpSame, "@1", "@1"},
		{CmpSame, "(rec :a int)", "(rec :a int)"},
		{CmpSame | BitWrap, "int", "int?"},
		{CmpSame | BitUnwrap, "int?", "int"},
		{CmpInfer, "int", "@"},
		{CmpCompAny, "int", "any"},
		{CmpCompList, "list|int", "~idxr"},
		{CmpCompList, "(rec :x :y int)", "~idxr"},
		{CmpCompDict, "dict|int", "~keyr"},
		{CmpCompDict, "(rec :x :y int)", "~keyr"},
		{CmpCompBase, "int", "~num"},
		{CmpCompBase | BitUnwrap, "int?", "~num"},
		{CmpCompSpec, "~num", "int"},
		{CmpCompSpec | BitWrap, "~num", "int?"},
		{CmpConvArr, "list|int?", "list|int"},
		{CmpConvMap, "dict|int?", "dict|int"},
		{CmpConvRec, "(rec :a int)", "(rec :a int?)"},
		{CmpConvRec, "(rec :foo :bar? int)", "(rec :foo int?)"},
		{CmpCheckRef, "@a", "int"},
		{CmpCheckRef, "int", "@a"},
		{CmpCheckAny, "any", "int"},
		{CmpCheckSpec, "~char", "uuid"},
		{CmpCheckSpec | BitUnwrap, "~char?", "uuid"},
		{CmpCheckSpec | BitWrap, "~char", "uuid?"},
		{CmpCheckList, "~idxr", "list|int"},
		{CmpCheckList, "~idxr", "(rec :x :y int)"},
		{CmpCheckDict, "~keyr", "dict|int"},
		{CmpCheckDict, "~keyr", "(rec :x :y int)"},
	}
	for _, test := range tests {
		s, err := ParseString(test.src)
		if err != nil {
			t.Errorf("parse src %s err: %v", test.src, err)
		}
		d, err := ParseString(test.dst)
		if err != nil {
			t.Errorf("parse dst %s err: %v", test.dst, err)
		}
		got := Compare(s, d)
		if got != test.want {
			t.Errorf("from %s to %s: want %v got %v", test.src, test.dst, test.want, got)
		}
	}
}
