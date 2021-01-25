package boltdb

type Config struct {
	// Dir is a directory to store the DB files
	Dir string
	// File is boltDB file. No need to create it by your own,
	// boltdb driver is able to create the file, or read existing
	File string
	// Bucket to store data in boltDB
	Bucket string
	// db file permissions
	Permissions int
	// timeout
	Interval uint `mapstructure:"interval"`
}

// InitDefaults initializes default values for the boltdb
func (s *Config) InitDefaults() {
	if s.Dir == "" {
		s.Dir = "." // current dir
	}
	if s.Bucket == "" {
		s.Bucket = "rr" // default bucket name
	}

	if s.File == "" {
		s.File = "rr.db" // default file name
	}

	if s.Permissions == 0 {
		s.Permissions = 777 // free for all
	}

	if s.Interval == 0 {
		s.Interval = 60 // default is 60 seconds timeout
	}
}
