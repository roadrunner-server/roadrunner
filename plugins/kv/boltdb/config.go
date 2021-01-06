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
	Interval uint `yaml:"interval"`
}

// InitDefaults initializes default values for the boltdb
func (s *Config) InitDefaults() {
	s.Dir = "."          // current dir
	s.Bucket = "rr"      // default bucket name
	s.File = "rr.db"     // default file name
	s.Permissions = 0777 // free for all
	s.Interval = 60      // default is 60 seconds timeout
}
