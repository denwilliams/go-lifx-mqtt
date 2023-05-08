package lifx

import "context"

type StatusEmitter interface {
	EmitStatus(ctx context.Context, id string, statusKey string, data interface{}) error
}
