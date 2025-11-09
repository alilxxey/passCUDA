package door

import (
	"context"
	"paSKUDa/internal/config"
)

type Door struct {
	ctx    context.Context
	config *config.Config
}

func Init(config *config.Config, ctx context.Context) (*Door, error) {
	return &Door{config: config, ctx: ctx}, nil
}
