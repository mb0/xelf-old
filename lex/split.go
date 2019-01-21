package lex

// Keyed has a list of trees named by a key
type Keyed struct {
	*Tree
	Key string
	Seq []*Tree
}

// SplitKeyed returns a list of keyed trees and also the head and tail list of trees
// It will check for symbols and symbols starting a nested expression.
// If uni is true each key can be followed by at most one trees otherwise by any number of trees
func SplitKeyed(seq []*Tree, uni bool, pred func(string) bool) (head, tail []*Tree, res []Keyed) {
	var last bool
	head = seq
	for i, t := range seq {
		if key, ok := CheckSym(t, 1, pred); ok {
			if len(res) == 0 {
				head = seq[:i]
				res = make([]Keyed, 0, len(seq)-i)
			}
			if IsExp(t) {
				res = append(res, Keyed{t, key, t.Seq[1:]})
				last = false
			} else {
				res = append(res, Keyed{t, key, nil})
				last = true
			}
		} else if last {
			res[len(res)-1].Seq = seq[i : i+1]
			last = !uni
		} else if len(res) > 0 {
			tail = seq[i:]
			break
		}
	}
	return
}

// Split splits seq at the first tree that matches pred and returns both head and tail
// If the first tree matches, it returns nil, seq; if no tree matches it returns seq, nil;
// otherwise it returns head without match and tail starting with a match
func Split(seq []*Tree, pred func(*Tree) bool) (head, tail []*Tree) {
	for i, t := range seq {
		if pred(t) {
			return seq[:i], seq[i:]
		}
	}
	return seq, nil
}

// SplitAfter splits seq and returns the head and tail after the matching tree
func SplitAfter(seq []*Tree, pred func(*Tree) bool) (head, tail []*Tree) {
	head, tail = Split(seq, pred)
	if len(tail) > 0 {
		return head, tail[1:]
	}
	return head, nil
}
