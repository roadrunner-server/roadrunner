package logger

// Mode represents available logger modes
type Mode string

const (
	none        Mode = "none"
	off         Mode = "off"
	production  Mode = "production"
	development Mode = "development"
	raw         Mode = "raw"
)
