package config

import (
	"sync"

	"github.com/hashicorp/hil"
	"github.com/hashicorp/hil/ast"
	"github.com/hashicorp/terraform/config/configschema"

	"github.com/mitchellh/copystructure"
	"github.com/mitchellh/reflectwalk"
)

type rawConfigSourceHCL struct {
	Key            string
	Raw            map[string]interface{}
	interpolations []ast.Node
	variables      map[string]InterpolatedVariable

	lock sync.Mutex
}

// NewRawConfig is a deprecated alias for NewRawConfigHCL.
func NewRawConfig(raw map[string]interface{}) (*RawConfig, error) {
	return NewRawConfigHCL(raw)
}

// NewRawConfigHCL creates a RawConfig from the result of decoding an
// HCL object into a map[string]interface{}.
//
// Any unescaped ${ ... } sequences within string values in the raw structure
// are interpreted as HIL interpolations.
func NewRawConfigHCL(raw map[string]interface{}) (*RawConfig, error) {
	src := &rawConfigSourceHCL{Raw: raw}
	if err := src.init(); err != nil {
		return nil, err
	}

	return &RawConfig{source: src}, nil
}

// NewRawConfigHCLKey is similar to NewRawConfigHCL except that its result
// is the final value of the given key in the result map, rather than the
// entire map.
func NewRawConfigHCLKey(raw map[string]interface{}, key string) (*RawConfig, error) {
	ret, err := NewRawConfigHCL(raw)
	if err != nil {
		return nil, err
	}

	ret.source.(*rawConfigSourceHCL).Key = key
	return ret, nil
}

func (r *rawConfigSourceHCL) Variables() map[string]InterpolatedVariable {
	r.lock.Lock()
	defer r.lock.Unlock()

	// shallow-copy our map so that the caller can mutate the return value
	// without corrupting our internals.
	ret := map[string]InterpolatedVariable{}
	for k, v := range r.variables {
		ret[k] = v
	}
	return ret
}

func (r *rawConfigSourceHCL) Interpolate(*configschema.Block, map[string]ast.Variable) (interface{}, []string, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	config := hilEvalConfig(vs)
	rmap, unks, err := r.interpolate(func(root ast.Node) (interface{}, error) {
		// None of the variables we need are computed, meaning we should
		// be able to properly evaluate.
		result, err := hil.Eval(root, config)
		if err != nil {
			return "", err
		}

		return result.Value, nil
	})

	if err != nil {
		return nil, nil, err
	}

	if r.Key != "" {
		return rmap[r.Key], unks, nil
	}

	return rmap, unks, nil
}

func (r *rawConfigSourceHCL) init() error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.config = r.Raw
	r.interpolations = nil
	r.variables = nil

	fn := func(node ast.Node) (interface{}, error) {
		r.interpolations = append(r.interpolations, node)
		vars, err := DetectVariables(node)
		if err != nil {
			return "", err
		}

		for _, v := range vars {
			if r.variables == nil {
				r.variables = make(map[string]InterpolatedVariable)
			}

			r.variables[v.FullKey()] = v
		}

		return "", nil
	}

	walker := &interpolationWalker{F: fn}
	if err := reflectwalk.Walk(r.Raw, walker); err != nil {
		return err
	}

	return nil
}

func (r *rawConfigSourceHCL) interpolate(fn interpolationWalkerFunc) (map[string]interface{}, []string, error) {
	rawCopy, err := copystructure.Copy(r.Raw)
	if err != nil {
		return err
	}
	config := rawCopy.(map[string]interface{})

	w := &interpolationWalker{F: fn, Replace: true}
	err = reflectwalk.Walk(config, w)
	return config, w.unknownKeys, err
}

// hilEvalConfig returns the evaluation configuration we use to execute.
func hilEvalConfig(vs map[string]ast.Variable) *hil.EvalConfig {
	funcMap := make(map[string]ast.Function)
	for k, v := range Funcs() {
		funcMap[k] = v
	}
	funcMap["lookup"] = interpolationFuncLookup(vs)
	funcMap["keys"] = interpolationFuncKeys(vs)
	funcMap["values"] = interpolationFuncValues(vs)

	return &hil.EvalConfig{
		GlobalScope: &ast.BasicScope{
			VarMap:  vs,
			FuncMap: funcMap,
		},
	}
}
