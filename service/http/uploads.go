package http

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

const (
	// UploadErrorOK - no error, the file uploaded with success.
	UploadErrorOK = 0

	// UploadErrorNoFile - no file was uploaded.
	UploadErrorNoFile = 4

	// UploadErrorNoTmpDir - missing a temporary folder.
	UploadErrorNoTmpDir = 5

	// UploadErrorCantWrite - failed to write file to disk.
	UploadErrorCantWrite = 6

	// UploadErrorExtension - forbidden file extension.
	UploadErrorExtension = 7
)

// Uploads tree manages uploaded files tree and temporary files.
type Uploads struct {
	// associated temp directory and forbidden extensions.
	cfg *UploadsConfig

	// pre processed data tree for Uploads.
	tree fileTree

	// flat list of all file Uploads.
	list []*FileUpload
}

// MarshalJSON marshal tree tree into JSON.
func (u *Uploads) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.tree)
}

// Open moves all uploaded files to temp directory, return error in case of issue with temp directory. File errors
// will be handled individually.
func (u *Uploads) Open(log *logrus.Logger) {
	var wg sync.WaitGroup
	for _, f := range u.list {
		wg.Add(1)
		go func(f *FileUpload) {
			defer wg.Done()
			err := f.Open(u.cfg)
			if err != nil && log != nil {
				log.Error(fmt.Errorf("error opening the file: error %v", err))
			}
		}(f)
	}

	wg.Wait()
}

// Clear deletes all temporary files.
func (u *Uploads) Clear(log *logrus.Logger) {
	for _, f := range u.list {
		if f.TempFilename != "" && exists(f.TempFilename) {
			err := os.Remove(f.TempFilename)
			if err != nil && log != nil {
				log.Error(fmt.Errorf("error removing the file: error %v", err))
			}
		}
	}
}

// FileUpload represents singular file NewUpload.
type FileUpload struct {
	// ID contains filename specified by the client.
	Name string `json:"name"`

	// Mime contains mime-type provided by the client.
	Mime string `json:"mime"`

	// Size of the uploaded file.
	Size int64 `json:"size"`

	// Error indicates file upload error (if any). See http://php.net/manual/en/features.file-upload.errors.php
	Error int `json:"error"`

	// TempFilename points to temporary file location.
	TempFilename string `json:"tmpName"`

	// associated file header
	header *multipart.FileHeader
}

// NewUpload wraps net/http upload into PRS-7 compatible structure.
func NewUpload(f *multipart.FileHeader) *FileUpload {
	return &FileUpload{
		Name:   f.Filename,
		Mime:   f.Header.Get("Content-Type"),
		Error:  UploadErrorOK,
		header: f,
	}
}

// Open moves file content into temporary file available for PHP.
// NOTE:
// There is 2 deferred functions, and in case of getting 2 errors from both functions
// error from close of temp file would be overwritten by error from the main file
// STACK
// DEFER FILE CLOSE (2)
// DEFER TMP CLOSE  (1)
func (f *FileUpload) Open(cfg *UploadsConfig) (err error) {
	if cfg.Forbids(f.Name) {
		f.Error = UploadErrorExtension
		return nil
	}

	file, err := f.header.Open()
	if err != nil {
		f.Error = UploadErrorNoFile
		return err
	}

	defer func() {
		// close the main file
		err = file.Close()
	}()

	tmp, err := ioutil.TempFile(cfg.TmpDir(), "upload")
	if err != nil {
		// most likely cause of this issue is missing tmp dir
		f.Error = UploadErrorNoTmpDir
		return err
	}

	f.TempFilename = tmp.Name()
	defer func() {
		// close the temp file
		err = tmp.Close()
	}()

	if f.Size, err = io.Copy(tmp, file); err != nil {
		f.Error = UploadErrorCantWrite
	}

	return err
}

// exists if file exists.
func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
