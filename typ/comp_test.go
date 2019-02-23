package typ

import "testing"

func TestComp(t *testing.T) {
	tests := []struct {
		want     Cmp
		src, dst string
	}{
		{CmpSame, "int", "int"},
		{CmpSame, "arr|int", "arr|int"},
		{CmpSame, "any", "any"},
		{CmpInfer, "any", "@"},
		{CmpCheckRef, "@ref", "@ref"},
		{CmpCheckRef, "@", "@"},
		{CmpSame, "(obj +a int)", "(obj +a int)"},
		{CmpSame | BitWrap, "int", "int?"},
		{CmpSame | BitUnwrap, "int?", "int"},
		{CmpInfer, "int", "@"},
		{CmpCompAny, "int", "any"},
		{CmpCompList, "arr|int", "list"},
		{CmpCompList, "(obj +x +y int)", "list"},
		{CmpCompDict, "map|int", "dict"},
		{CmpCompDict, "(obj +x +y int)", "dict"},
		{CmpCompBase, "int", "num"},
		{CmpCompBase | BitUnwrap, "int?", "num"},
		{CmpCompSpec, "num", "int"},
		{CmpCompSpec | BitWrap, "num", "int?"},
		{CmpConvArr, "arr|int?", "arr|int"},
		{CmpConvMap, "map|int?", "map|int"},
		{CmpConvObj, "(obj +a int)", "(obj +a int?)"},
		{CmpConvObj, "(obj +foo +bar? int)", "(obj +foo int?)"},
		{CmpCheckRef, "@a", "int"},
		{CmpCheckRef, "int", "@a"},
		{CmpCheckAny, "any", "int"},
		{CmpCheckSpec, "char", "uuid"},
		{CmpCheckSpec | BitUnwrap, "char?", "uuid"},
		{CmpCheckSpec | BitWrap, "char", "uuid?"},
		{CmpCheckList, "list", "arr|int"},
		{CmpCheckList, "list", "(obj +x +y int)"},
		{CmpCheckDict, "dict", "map|int"},
		{CmpCheckDict, "dict", "(obj +x +y int)"},
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
