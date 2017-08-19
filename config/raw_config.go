package config

import (
	"github.com/hashicorp/hil/ast"
	"github.com/hashicorp/terraform/config/configschema"
)

// UnknownVariableValue is a sentinel value that can be used
// to denote that the value of a variable is unknown at this time.
// RawConfig uses this information to build up data about
// unknown keys.
const UnknownVariableValue = "74D93920-ED26-11E3-AC10-0800200C9A66"

// RawConfig is a structure that holds a piece of configuration
// where the overall structure is unknown since it will be used
// to configure a plugin or some other similar external component.
//
// RawConfigs can be interpolated with variables that come from
// other resources, user variables, etc.
type RawConfig struct {
	source rawConfigSource
}

// rawConfigSource is a type implemented for each supported configuration
// language that knows how to deal with that language's configuration blocks.
type rawConfigSource interface {
	Variables() map[string]InterpolatedVariable
	Interpolate(*configschema.Block, map[string]ast.Variable) (result map[string]interface{}, unknownKeys []string, err error)
}

func (r *RawConfig) Variables() map[string]InterpolatedVariable {
	return r.source.Variables()
}

func (r *RawConfig) Interpolate(schema *configschema.Block, variables map[string]ast.Variable) (result interface{}, unknownKeys []string, err error) {
	return r.source.Interpolate(schema, variables)
}

// Merge merges another RawConfig into this one (overriding any conflicting
// values in this config) and returns a new config. The original config
// is not modified.
func (r *RawConfig) Merge(other *RawConfig) *RawConfig {
	return MergeRawConfigs(r, other)
}
