package boltjobs

type Config struct {
	// File is boltDB file. No need to create it by your own,
	// boltdb driver is able to create the file, or read existing
	File string
	// Bucket to store data in boltDB
	bucket string
	// db file permissions
	Permissions int
	// consume timeout
}

func (c *Config) InitDefaults() {

}
