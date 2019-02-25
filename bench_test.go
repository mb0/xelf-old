package xelf

import "testing"

type env interface {
	support(byte) bool
}

type supportX interface {
	supportX()
}

type myenv struct {
	data []byte
}

func (e *myenv) support(c byte) bool { return c == 'x' }
func (e *myenv) supportX()           {}

func newEnv() env {
	return new(myenv)
}

func BenchmarkCall(b *testing.B) {
	e, res := newEnv(), 0
	for i := 0; i < b.N; i++ {
		if e.support('x') {
			res++
		}
	}
}

func BenchmarkMarkerConv(b *testing.B) {
	e, res := newEnv(), 0
	for i := 0; i < b.N; i++ {
		if _, ok := e.(supportX); ok {
			res++
		}
	}
}
func BenchmarkMarkerSwitch(b *testing.B) {
	e, res := newEnv(), 0
	for i := 0; i < b.N; i++ {
		switch e.(type) {
		case supportX:
			res++
		}
	}
}
