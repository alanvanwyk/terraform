package config

import (
	"github.com/hashicorp/hil/ast"
	"github.com/hashicorp/terraform/config/configschema"
)

// MergeRawConfigs returns a new RawConfig that behaves as if it contained
// the contents of both of the given RawConfigs.
//
// The configurations are merged only in a shallow way; nested lists and maps
// are not merged. If both given RawConfigs define the same key, the value
// from RawConfig b takes precendence.
func MergeRawConfigs(a *RawConfig, b *RawConfig) *RawConfig {
	return &RawConfig{
		source: &rawConfigSourceMerged{
			a: a.source,
			b: b.source,
		},
	}
}

type rawConfigSourceMerged struct {
	a rawConfigSource
	b rawConfigSource
}

func (r *rawConfigSourceMerged) Variables() map[string]InterpolatedVariable {
	ret := map[string]InterpolatedVariable{}

	va := r.a.Variables()
	vb := r.b.Variables()

	// If vb overrides an expression in va, we may actually produce references
	// to variables that will ultimately not be used. This is accepted because
	// it's the more conservative path.
	for k, v := range va {
		ret[k] = v
	}
	for k, v := range vb {
		ret[k] = v
	}

	return ret
}

func (r *rawConfigSourceMerged) Interpolate(schema *configschema.Block, vs map[string]ast.Variable) (interface{}, []string, error) {

	ra, unka, err := r.a.Interpolate(schema, vs)
	if err != nil {
		return nil, nil, err
	}
	rb, unkb, err := r.b.Interpolate(schema, vs)
	if err != nil {
		return nil, nil, err
	}

	// If result b isn't a map then no merging is possible and we'll just
	// return result b.
	if _, isMap := rb.(map[string]interface{}); !isMap {
		return rb, unkb, nil
	}

	result := map[string]interface{}{}
	for k, v := range ra.(map[string]interface{}) {
		result[k] = v
	}
	for k, v := range rb.(map[string]interface{}) {
		result[k] = v
	}

	unkMap := map[string]struct{}{}
	var unk []string

	// This result actually over-reports unknowns, since something might
	// be unknown in a but not in b, with b taking precedence. This was,
	// however, the behavior of the former "merge" implementation on
	// the old HCL-only RawConfig, so it's preserved here for compatibility.
	for _, k := range unka {
		unks[k] = struct{}{}
	}
	for _, k := range unkb {
		unks[k] = struct{}{}
	}

	if len(unkMap) > 0 {
		unk = make([]string, 0, len(unkMap))
		for k := range unks {
			unk = append(unk, k)
		}
	}

	return result, unk, nil
}
