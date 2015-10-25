package api

import (
	"github.com/hashicorp/terraform/terraform"
)

type InstanceStateMessage struct {
	ID         string                        `json:"id"`
	Attributes map[string]string             `json:"attributes"`
	Ephemeral  InstanceStateEphemeralMessage `json:"ephemeral"`
	Meta       map[string]string             `json:"meta"`
}

func NewInstanceStateMessage(s *terraform.InstanceState) *InstanceStateMessage {
	return &InstanceStateMessage{
		ID:         s.ID,
		Attributes: s.Attributes,
		Ephemeral: InstanceStateEphemeralMessage{
			ConnInfo: s.Ephemeral.ConnInfo,
		},
		Meta: s.Meta,
	}
}

func (i *InstanceStateMessage) InstanceState() *terraform.InstanceState {
	return &terraform.InstanceState{
		ID:         i.ID,
		Attributes: i.Attributes,
		//Ephemeral: i.Ephemeral.EphemeralState(),
		Meta: i.Meta,
	}
}

type InstanceStateEphemeralMessage struct {
	ConnInfo map[string]string `json:"connectionInfo"`
}

func (i *InstanceStateEphemeralMessage) EphemeralState() terraform.EphemeralState {
	return terraform.EphemeralState{
		ConnInfo: i.ConnInfo,
	}
}
