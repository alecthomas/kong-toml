package kongtoml

import (
	"io"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/repr"
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
	repr.Println(tree.ToMap())
	return &Resolver{filename: filename, tree: tree.ToMap()}, nil
}

var _ kong.Resolver = (*Resolver)(nil)

type Resolver struct {
	filename string
	tree     map[string]any
}

func (r *Resolver) Resolve(kctx *kong.Context, parent *kong.Path, flag *kong.Flag) (any, error) {
	value, ok := r.findValue(parent, flag)
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (r *Resolver) Validate(app *kong.Application) error {
	return nil
}

func (r *Resolver) findValue(parent *kong.Path, flag *kong.Flag) (any, bool) {
	keys := []string{
		strings.Join(append(strings.Split(parent.Node().Path(), "-"), flag.Name), "-"),
		flag.Name,
	}
	for _, key := range keys {
		parts := strings.Split(key, "-")
		if value, ok := r.findValueParts(parts[0], parts[1:], r.tree); ok {
			return value, ok
		}
	}
	return nil, false
}

func (r *Resolver) findValueParts(prefix string, suffix []string, tree map[string]any) (any, bool) {
	if value, ok := tree[prefix]; ok {
		if len(suffix) == 0 {
			return value, true
		}
		if branch, ok := value.(map[string]any); ok {
			return r.findValueParts(suffix[0], suffix[1:], branch)
		}
	}
	if len(suffix) > 0 {
		return r.findValueParts(prefix+"-"+suffix[0], suffix[1:], tree)
	}
	return nil, false
}
