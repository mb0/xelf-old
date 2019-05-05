package typ

import (
	"github.com/mb0/xelf/cor"
)

// Ctx is used to check and infer type variables.
type Ctx struct {
	binds Binds
	last  uint64
}

// New returns a new type variable for this context.
func (c *Ctx) New() Type {
	c.last++
	return Var(uint64(c.last))
}

// Bind binds type variable v to type t or returns an error.
func (c *Ctx) Bind(v Kind, t Type) error {
	if v&MaskRef != KindVar {
		return cor.Errorf("not a type variable %s", v)
	}
	id := uint64(v >> SlotSize)
	if id == 0 {
		return cor.Errorf("type variable without id %s", v)
	}
	if id > c.last {
		c.last = id
	}
	if c.Contains(t, v) {
		return cor.Errorf("recursive type variable %s", v)
	}
	c.binds = c.binds.Set(v, t)
	return nil
}

// Apply returns t with variables replaced from context.
func (c *Ctx) Apply(t Type) Type { t, _ = c.apply(t); return Choose(c, t) }
func (c *Ctx) apply(t Type) (s Type, ok bool) {
	for isVar(t) {
		if s, ok = c.binds.Get(t.Kind); ok {
			t = s
			continue
		}
		break
	}
	if !t.HasParams() {
		return t, ok
	}
	var ps []Param
	for i, p := range t.Params {
		pt, ok := c.apply(p.Type)
		if ok && ps == nil {
			ps = make([]Param, i, len(t.Params))
			copy(ps, t.Params)
		}
		if ps != nil {
			p.Type = pt
			ps = append(ps, p)
		}
	}
	if ps != nil {
		n := *t.Info
		n.Params = ps
		return Type{t.Kind, &n}, true
	}
	return t, ok
}

// Inst instantiates type t for this context, replacing all type vars.
func (c *Ctx) Inst(t Type) Type { r, _ := c.inst(t, nil); return r }
func (c *Ctx) inst(t Type, m Binds) (Type, Binds) {
	t, _ = c.apply(t)
	if isVar(t) {
		r, ok := m.Get(t.Kind)
		if !ok {
			r = c.New()
			r.Info = t.Info
			m = m.Set(t.Kind, r)
		} else if t.HasParams() {
			r = mergeHint(r, t)
			return r, m
		}
		return r, m
	} else if t.HasParams() {
		n := *t.Info
		r := Type{Kind: t.Kind, Info: &n}
		r.Params = make([]Param, 0, len(t.Params))
		for _, p := range t.Params {
			p.Type, m = c.inst(p.Type, m)
			r.Params = append(r.Params, p)
		}
		return r, m
	}
	return t, m
}

// Bound returns vars with all type variables in t, that are bound to this context, appended.
func (c *Ctx) Bound(t Type, vars Vars) Vars {
	if isVar(t) {
		if _, ok := c.binds.Get(t.Kind); ok {
			vars = vars.Add(t.Kind)
		}
	} else if t.HasParams() {
		for _, p := range t.Params {
			vars = c.Bound(p.Type, vars)
		}
	}
	return vars
}

// Free returns vars with all unbound type variables in t appended.
func (c *Ctx) Free(t Type, vars Vars) Vars {
	if isVar(t) {
		if r, ok := c.binds.Get(t.Kind); ok {
			vars = c.Free(r, vars)
			vars = vars.Del(t.Kind)
		} else {
			vars = vars.Add(t.Kind)
		}
	} else if t.HasParams() {
		for _, p := range t.Params {
			vars = c.Free(p.Type, vars)
		}
	}
	return vars
}

// Contains returns whether t contains the type variable v.
func (c *Ctx) Contains(t Type, v Kind) bool {
	for {
		if isVar(t) {
			if t.Kind == v {
				return true
			}
			t, _ = c.binds.Get(t.Kind)
			continue
		}
		if t.HasParams() {
			for _, p := range t.Params {
				if c.Contains(p.Type, v) {
					return true
				}
			}
		}
		return false
	}
}
