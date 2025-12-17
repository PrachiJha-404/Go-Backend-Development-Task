package middleware

import(
	"time"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var logger *zap.Logger
func SetLogger(l *zap.Logger){
	logger = l
}

func RequestLogger() fiber.Handler{
	return func (c *fiber.Ctx) error{
		start:= time.Now()
		err := c.Next()
		duration := time.Since(start)
		logger.Info("HTTP Request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("duration", duration),
			zap.String("ip", c.IP()),
			zap.String("user_agent", c.Get("User-Agent")),
)
		return err
	}
}

func ErrorHandler() fiber.Handler{
	return func(c* fiber.Ctx) error{
		err := c.Next()
		if err!=nil{
			logger.Error("Request error",
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.Error(err),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":"internal server error",
			})
		}
		return nil
	}
}

func CORS() fiber.Handler{
	return func(c *fiber.Ctx) error{
		c.Set("Access-Control-Allow-Origin", "*")
		c.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Method() == "OPTIONS"{
			return c.SendStatus(fiber.StatusNoContent)
		}
		return c.Next()
	}
}