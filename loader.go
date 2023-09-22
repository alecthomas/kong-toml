package kongtoml

import (
	"io"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/pelletier/go-toml"
)

func Loader(r io.Reader) (kong.Resolver, error) {
	tree, err := toml.LoadReader(r)
	if err != nil {
		return nil, err
	}
	var filename string
	if named, ok := r.(interface{ Name() string }); ok {
		filename = named.Name()
	}
	return &Resolver{filename: filename, tree: tree.ToMap()}, nil
}

var _ kong.Resolver = (*Resolver)(nil)

type Resolver struct {
	filename string
	tree     map[string]any
}

func (r *Resolver) Resolve(kctx *kong.Context, parent *kong.Path, flag *kong.Flag) (interface{}, error) {
	value, ok := r.findValue(flag)
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (r *Resolver) Validate(app *kong.Application) error {
	// TODO: Validate the configuration maps to valid flags.
	return nil
}

func (r *Resolver) findValue(flag *kong.Flag) (any, bool) {
	parts := strings.Split(flag.Name, "-")
	return r.findValueParts(parts[0], parts[1:], r.tree)
}

func (r *Resolver) findValueParts(prefix string, suffix []string, tree map[string]any) (any, bool) {
	if value, ok := r.tree[prefix]; ok {
		if len(suffix) == 0 {
			return value, true
		}
		if branch, ok := value.(map[string]any); ok {
			return r.findValueParts(prefix+"-"+suffix[0], suffix[1:], branch)
		}
	}
	return nil, false
}
