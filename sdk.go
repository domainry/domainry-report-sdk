package reportsdk

import "context"

const ProtocolVersionV1 = "domainry-report-protocol-v1"

type ApplicationRef struct{ RuntimeID string }

type Descriptor struct {
	ProtocolVersion string
	Mode            string
}

type Binding interface {
	Descriptor() Descriptor
	Close(context.Context) error
}
