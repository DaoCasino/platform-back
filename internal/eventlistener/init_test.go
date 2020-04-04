package eventlistener

import (
	"context"
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.With().Str("package", "EventListener").Bool("test", true).Logger()
	ctx := log.Logger.WithContext(context.Background())
	logger = log.Ctx(ctx)
}
