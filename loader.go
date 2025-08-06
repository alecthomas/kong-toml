package kongtoml

import (
	"fmt"
	"io"
	"maps"
	"slices"
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

func (r *Resolver) Resolve(kctx *kong.Context, parent *kong.Path, flag *kong.Flag) (any, error) {
	value, ok := r.findValue(parent, flag)
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (r *Resolver) Validate(app *kong.Application) error {
	configKeys := map[string]bool{}
	flattenTree("", r.tree, configKeys)
	_ = kong.Visit(app, func(node kong.Visitable, next kong.Next) error {
		if flag, ok := node.(*kong.Flag); ok {
			delete(configKeys, flag.Name)
		}
		return next(nil)
	})
	if len(configKeys) > 0 {
		keys := slices.Collect(maps.Keys(configKeys))
		return fmt.Errorf("%s: unknown configuration keys: %s", r.filename, strings.Join(keys, ", "))
	}
	return nil
}

func flattenTree(prefix string, tree any, flags map[string]bool) {
	switch tree := tree.(type) {
	case map[string]any:
		for key, value := range tree {
			if prefix == "" {
				flattenTree(key, value, flags)
			} else {
				flattenTree(prefix+"-"+key, value, flags)
			}
		}
	default:
		flags[prefix] = true
	}
}

func (r *Resolver) findValue(parent *kong.Path, flag *kong.Flag) (any, bool) {
	keys := []string{
		strings.Join(append(strings.Split(parent.Node().Path(), "-"), flag.Name), "-"),
		flag.Name,
	}
	return r.findValueFromKeys(keys)
}

func (r *Resolver) findValueFromKeys(keys []string) (any, bool) {
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
