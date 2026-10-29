package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	JWTSecret      string
	JWTExpiryHours int

	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	BookingTimezone    string // IANA zone the working hours are defined in
	BookingDayStart    string // "HH:MM"
	BookingDayEnd      string // "HH:MM"
	BookingWorkDays    string // comma-separated, e.g. "sun,mon,tue,wed,thu"
	BookingSlotMinutes int
}

func Load() *Config {
	// Ignore the error: in Docker the vars come from the environment.
	_ = godotenv.Load()

	expiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		log.Fatalf("invalid JWT_EXPIRY_HOURS: %v", err)
	}

	slotMinutes, err := strconv.Atoi(getEnv("BOOKING_SLOT_MINUTES", "30"))
	if err != nil {
		log.Fatalf("invalid BOOKING_SLOT_MINUTES: %v", err)
	}

	return &Config{
		AppPort: getEnv("APP_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "calendar_booking"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		JWTSecret:      mustGetEnv("JWT_SECRET"),
		JWTExpiryHours: expiry,



		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/google/callback"),


		BookingTimezone:    getEnv("BOOKING_TIMEZONE", "Africa/Cairo"),
		BookingDayStart:    getEnv("BOOKING_DAY_START", "09:00"),
		BookingDayEnd:      getEnv("BOOKING_DAY_END", "17:00"),
		BookingWorkDays:    getEnv("BOOKING_WORK_DAYS", "sun,mon,tue,wed,thu"),
		BookingSlotMinutes: slotMinutes,





	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}