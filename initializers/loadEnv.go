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

	// Cloudflare R2 (S3 API) for image uploads; all four set = uploads enabled.
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2Bucket          string
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

		R2AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2Bucket:          os.Getenv("R2_BUCKET"),
	}, nil
}

// HasR2 reports whether image uploads are configured.
func (c *Config) HasR2() bool {
	return c.R2AccountID != "" && c.R2AccessKeyID != "" && c.R2SecretAccessKey != "" && c.R2Bucket != ""
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
