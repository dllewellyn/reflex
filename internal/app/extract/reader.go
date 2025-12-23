package extract

import (
	"context"
)

type ResultReader interface {
	ReadResults(ctx context.Context) (<-chan BatchResult, <-chan error, func())
}
