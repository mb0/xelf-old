package typ

import (
	"strings"
	"testing"
)

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
		{CmpSame, "@", "@"},
		{CmpInfer, "@", "@1"},
		{CmpSame, "<rec a:int>", "<rec a:int>"},
		{CmpSame | BitWrap, "int", "int?"},
		{CmpSame | BitUnwrap, "int?", "int"},
		{CmpInfer, "int", "@"},
		{CmpCompAny, "int", "any"},
		{CmpCompList, "list|int", "list"},
		{CmpCompList, "<rec x:int y:int>", "list"},
		{CmpCompDict, "dict|int", "dict"},
		{CmpCompDict, "<rec x:int y:int>", "dict"},
		{CmpCompBase, "int", "num"},
		{CmpCompBase | BitUnwrap, "int?", "num"},
		{CmpCompSpec, "num", "int"},
		{CmpCompSpec, "num", "span"},
		{CmpCompSpec, "num", "time"},
		{CmpCheckSpec, "char", "span"},
		{CmpCheckSpec, "char", "time"},
		{CmpCompSpec | BitWrap, "num", "int?"},
		{CmpConvList, "list|int?", "list|int"},
		{CmpConvDict, "dict|int?", "dict|int"},
		{CmpConvRec, "<rec a:int>", "<rec a:int?>"},
		{CmpConvRec, "<rec foo:int bar?:int>", "<rec foo:int?>"},
		{CmpCheckRef, "@a", "int"},
		{CmpCheckRef, "int", "@a"},
		{CmpCheckAny, "any", "int"},
		{CmpCheckSpec, "char", "uuid"},
		{CmpCheckSpec | BitUnwrap, "char?", "uuid"},
		{CmpCheckSpec | BitWrap, "char", "uuid?"},
		{CmpCheckListAny, "list", "list|int"},
		{CmpCheckListAny, "list", "<rec x:int y:int>"},
		{CmpCheckDictAny, "dict", "dict|int"},
		{CmpCheckDictAny, "dict", "<rec x:int y:int>"},
	}
	for _, test := range tests {
		s, err := Read(strings.NewReader(test.src))
		if err != nil {
			t.Errorf("parse src %s err: %v", test.src, err)
		}
		d, err := Read(strings.NewReader(test.dst))
		if err != nil {
			t.Errorf("parse dst %s err: %v", test.dst, err)
		}
		got := Compare(s, d)
		if got != test.want {
			t.Errorf("from %s to %s: want %v got %v", test.src, test.dst, test.want, got)
		}
	}
}
