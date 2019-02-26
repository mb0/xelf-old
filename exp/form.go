package exp

import (
	"github.com/mb0/xelf/bfr"
	"github.com/mb0/xelf/cor"
	"github.com/mb0/xelf/lex"
	"github.com/mb0/xelf/typ"
)

// Form represents a from resolver and implements both literal and resolver interface.
type Form struct {
	Sig  Type
	Rslv Resolver
}

func (f *Form) Typ() typ.Type  { return f.Sig }
func (f *Form) IsZero() bool   { return false }
func (f *Form) String() string { return bfr.String(f) }
func (f *Form) WriteBfr(b bfr.Ctx) error {
	b.WriteByte('(')
	err := f.Sig.WriteBfr(b)
	if err != nil {
		return err
	}
	return b.WriteByte(')')
}

func (f *Form) MarshalJSON() ([]byte, error) {
	v, err := lex.Quote(f.String(), '"')
	if err != nil {
		return nil, err
	}
	return []byte(v), nil
}

func (f *Form) Resolve(c *Ctx, env Env, e El, hint Type) (El, error) {
	switch e.(type) {
	case *Ref:
		return f, nil
	case *Expr:
		return f.Rslv.Resolve(c, env, e, hint)
	}
	return nil, cor.Errorf("unexpected element %T", e)
}
