package global

import (
	"fmt"

	"go.uber.org/zap"
)

var Log *zap.Logger

func InitLogger() error {
	l, err := zap.NewDevelopment()
	if err != nil {
		//往上抛 error
		return fmt.Errorf("init logger: %w", err)
	}
	Log = l
	return nil
}
