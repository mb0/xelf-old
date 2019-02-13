// Package gen provides code generation helpers and specific functions to generate go code.
package gen

import (
	"sort"

	"github.com/mb0/xelf/bfr"
)

// Ctx is the code generation context holding the buffer and addition information.
type Ctx struct {
	bfr.B
	Pkg    string
	Target string
	Header string
	Pkgs   map[string]string
	Imports
}

// Imports has a list of alphabetically sorted dependencies. A dependency can be any string
// recognized by the generator. For go imports the dependency is a package path.
type Imports struct {
	List []string
}

// Add inserts path into the import list if not already present.
func (i *Imports) Add(path string) {
	idx := sort.SearchStrings(i.List, path)
	if idx < len(i.List) && i.List[idx] == path {
		return
	}
	i.List = append(i.List, "")
	copy(i.List[idx+1:], i.List[idx:])
	i.List[idx] = path
}
