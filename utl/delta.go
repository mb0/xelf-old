package utl

import (
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lit"
	"github.com/mb0/xelf/typ"
)

func NewDelta(a, b lit.Obj) (*lit.Dict, error) {
	t := a.Typ()
	cmp := typ.Compare(t, b.Typ())
	if cmp < typ.LvlComp {
		return nil, cor.Errorf("types not comparable %s %s", t, b.Typ())
	}
	res := &lit.Dict{}
	for _, p := range t.Params {
		key := p.Key()
		av, err := a.Key(key)
		if err != nil {
			return nil, err
		}
		bv, err := b.Key(key)
		if err != nil {
			return nil, err
		}
		if lit.Equal(av, bv) {
			continue
		}
		// TODO check if container or char and diff
		res.List = append(res.List, lit.Keyed{key, bv})
	}
	return res, nil
}

func MergeDeltas(a, b *lit.Dict) error {
	for _, kv := range b.List {
		// TODO check for common prefix, but order preserving dict works for now
		_, err := a.SetKey(kv.Key, kv.Lit)
		if err != nil {
			return err
		}
	}
	return nil
}

func ApplyDelta(o lit.Keyer, d *lit.Dict) error {
	for _, kv := range d.List {
		p, err := lit.ReadPath(kv.Key)
		if err != nil {
			return err
		}
		_, err = lit.SetPath(o, p, kv.Lit, true)
		if err != nil {
			return err
		}
	}
	return nil
}
