package deps

import (
	"fmt"
	"os"

	"github.com/blendle/zapdriver"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger() (*zap.SugaredLogger, error) {
	var config zap.Config

	config = zapdriver.NewProductionConfig()

	if os.Getenv("ON_GCP") != "true" {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("zap.Build() failed: %v", err)
	}

	return logger.Sugar(), nil
}
