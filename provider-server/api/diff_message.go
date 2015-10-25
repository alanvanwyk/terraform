package api

import (
	"fmt"

	"github.com/hashicorp/terraform/terraform"
)

type DiffMessage struct {
	Attributes     map[string]DiffMessageAttr `json:"attributes"`
	Destroy        bool                       `json:"destroy"`
	DestroyTainted bool                       `json:"destroyTainted"`
}

type DiffMessageAttr struct {
	Old         string              `json:"oldValue"`
	New         string              `json:"newValue"`
	NewComputed bool                `json:"newIsComputed"`
	NewRemoved  bool                `json:"newIsRemoved"`
	NewExtra    interface{}         `json:"newExtra,omitempty"`
	RequiresNew bool                `json:"requiresNew"`
	Type        DiffMessageAttrType `json:"type"`
}

type DiffMessageAttrType string

const (
	DiffMessageAttrUnknown DiffMessageAttrType = "unknown"
	DiffMessageAttrInput   DiffMessageAttrType = "input"
	DiffMessageAttrOutput  DiffMessageAttrType = "output"
)

func NewDiffMessage(diff *terraform.InstanceDiff) *DiffMessage {
	ret := &DiffMessage{
		Attributes:     make(map[string]DiffMessageAttr),
		Destroy:        diff.Destroy,
		DestroyTainted: diff.DestroyTainted,
	}
	for k, attrDiff := range diff.Attributes {
		ret.Attributes[k] = *NewDiffMessageAttr(attrDiff)
	}
	return ret
}

func (m *DiffMessage) InstanceDiff() *terraform.InstanceDiff {
	ret := &terraform.InstanceDiff{
		Attributes:     make(map[string]*terraform.ResourceAttrDiff),
		Destroy:        m.Destroy,
		DestroyTainted: m.DestroyTainted,
	}
	for k, attrDiff := range m.Attributes {
		ret.Attributes[k] = attrDiff.ResourceAttrDiff()
	}
	return ret
}

func NewDiffMessageAttr(attrDiff *terraform.ResourceAttrDiff) *DiffMessageAttr {
	return &DiffMessageAttr{
		Old:         attrDiff.Old,
		New:         attrDiff.New,
		NewComputed: attrDiff.NewComputed,
		NewRemoved:  attrDiff.NewRemoved,
		NewExtra:    attrDiff.NewExtra,
		RequiresNew: attrDiff.RequiresNew,
		Type:        MakeMessageAttrType(attrDiff.Type),
	}
}

func (m *DiffMessageAttr) ResourceAttrDiff() *terraform.ResourceAttrDiff {
	return &terraform.ResourceAttrDiff{
		Old:         m.Old,
		New:         m.New,
		NewComputed: m.NewComputed,
		NewRemoved:  m.NewRemoved,
		NewExtra:    m.NewExtra,
		RequiresNew: m.RequiresNew,
		Type:        m.Type.DiffAttrType(),
	}
}

func MakeMessageAttrType(t terraform.DiffAttrType) DiffMessageAttrType {
	switch t {
	case terraform.DiffAttrUnknown:
		return DiffMessageAttrUnknown
	case terraform.DiffAttrInput:
		return DiffMessageAttrInput
	case terraform.DiffAttrOutput:
		return DiffMessageAttrOutput
	default:
		panic(fmt.Errorf("Unknown message attr type %v", t))
	}
}

func (t DiffMessageAttrType) DiffAttrType() (terraform.DiffAttrType) {
	switch t {
	case DiffMessageAttrInput:
		return terraform.DiffAttrInput
	case DiffMessageAttrOutput:
		return terraform.DiffAttrOutput
	default:
		// Be liberal in what we accept
		return terraform.DiffAttrUnknown
	}
}
