package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	playgroundShareURL = "https://play.golang.org/share"
	playgroundBaseURL  = "https://play.golang.org/p/"
)

type Config struct {
	Port     string `mapstructure:"PORT"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

func main() {
	// Initialize configuration
	cfg := &Config{
		Port:     "8080",
		LogLevel: "info",
	}

	// Configure Viper with experimental struct binding
	viper.SetOptions(viper.ExperimentalBindStruct())
	viper.SetEnvPrefix("GOPLAY")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind the struct
	if err := viper.Unmarshal(cfg); err != nil {
		fmt.Printf("Error unmarshaling config: %v\n", err)
		os.Exit(1)
	}

	// Initialize Zap logger
	logger, err := initLogger(cfg.LogLevel)
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			logger.Error("Error syncing logger", zap.Error(err))
		}
	}(logger)

	// Create Echo instance
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogError:     true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogUserAgent: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			logger.Info("request",
				zap.String("uri", v.URI),
				zap.Int("status", v.Status),
				zap.String("method", v.Method),
				zap.String("remote_ip", v.RemoteIP),
				zap.String("user_agent", v.UserAgent),
				zap.Duration("latency", v.Latency),
				zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
			)
			if v.Error != nil {
				logger.Error("request error", zap.Error(v.Error))
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())

	// Routes
	e.GET("/health", handleHealth)
	e.GET("/", makeHandleProxy(logger))

	// Start server
	port := cfg.Port
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	logger.Info("Go Playground Proxy starting",
		zap.String("port", port),
		zap.String("log_level", cfg.LogLevel),
	)
	logger.Info("Usage", zap.String("url", fmt.Sprintf("http://localhost%s/?code={code}", port)))

	if err := e.Start(port); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

func initLogger(level string) (*zap.Logger, error) {
	// Parse log level
	var zapLevel zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Create logger configuration
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      zapLevel == zapcore.DebugLevel,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// Use console encoding for development
	if config.Development {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return config.Build()
}

func handleHealth(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func makeHandleProxy(logger *zap.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get the code parameter
		code := c.QueryParam("code")
		if code == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Missing 'code' parameter",
			})
		}

		// URL decode the code
		decodedCode, err := url.QueryUnescape(code)
		if err != nil {
			logger.Error("Error decoding code", zap.Error(err))
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Invalid code parameter",
			})
		}

		// Share the code with the official Go Playground
		shareID, err := shareWithPlayground(logger, decodedCode)
		if err != nil {
			logger.Error("Error sharing with playground", zap.Error(err))
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to share code with playground",
			})
		}

		// Redirect to the playground URL
		playgroundURL := playgroundBaseURL + shareID
		logger.Info("Redirecting to playground",
			zap.String("url", playgroundURL),
			zap.String("share_id", shareID),
		)
		return c.Redirect(http.StatusFound, playgroundURL)
	}
}

func shareWithPlayground(logger *zap.Logger, code string) (string, error) {
	// Create a POST request to share the code
	resp, err := http.Post(playgroundShareURL, "application/x-www-form-urlencoded", strings.NewReader(code))
	if err != nil {
		return "", fmt.Errorf("failed to post to playground: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("Error closing response body", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("playground returned status %d", resp.StatusCode)
	}

	// Read the share ID from the response
	shareID, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	shareIDStr := strings.TrimSpace(string(shareID))
	if shareIDStr == "" {
		return "", fmt.Errorf("empty share ID received")
	}

	return shareIDStr, nil
}
