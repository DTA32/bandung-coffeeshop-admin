package initializers

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// LoadEnv reads a local .env file if present (ignored in containers).
func LoadEnv() { _ = godotenv.Load() }

type Config struct {
	AppPort      string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	ImageBaseURL string
	DisplayTZ    *time.Location
}

func LoadConfig() (*Config, error) {
	tzName := getenv("DISPLAY_TZ", "Asia/Jakarta")
	tz, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("DISPLAY_TZ: %w", err)
	}
	return &Config{
		AppPort:      getenv("APP_PORT", "8080"),
		DBHost:       os.Getenv("DB_HOST"),
		DBPort:       getenv("DB_PORT", "5432"),
		DBUser:       os.Getenv("DB_USER"),
		DBPassword:   os.Getenv("DB_PASSWORD"),
		DBName:       os.Getenv("DB_NAME"),
		DBSSLMode:    getenv("DB_SSLMODE", "disable"),
		ImageBaseURL: strings.TrimRight(getenv("IMAGE_BASE_URL", "https://image.bdgcafe.com"), "/"),
		DisplayTZ:    tz,
	}, nil
}

// DSN builds a postgres URL. The session timezone is pinned to UTC so that
// `timestamp` columns written via now() and values written by this app agree.
func (c *Config) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DBUser, c.DBPassword),
		Host:   c.DBHost + ":" + c.DBPort,
		Path:   "/" + c.DBName,
	}
	q := url.Values{}
	q.Set("sslmode", c.DBSSLMode)
	q.Set("timezone", "UTC")
	u.RawQuery = q.Encode()
	return u.String()
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
