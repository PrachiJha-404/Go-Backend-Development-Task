package logger
import(
	"os"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(env string) (*zap.Logger, error){
	var config zap.Config

	if env=="production"{
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey="timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else{
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	logger, err := config.Build()
	if err != nil{
		return nil, err
	}
	return logger, nil
}

func NewLoggerFromEnv() (*zap.Logger, error){
	env := os.Getenv("APP_ENV")
	if env==""{
		env = "development"
	}
	return NewLogger(env)
}