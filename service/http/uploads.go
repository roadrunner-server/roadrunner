package http

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"sync"
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
func (u *Uploads) Open() {
	var wg sync.WaitGroup
	for _, f := range u.list {
		wg.Add(1)
		go func(f *FileUpload) {
			defer wg.Done()
			f.Open(u.cfg)
		}(f)
	}

	wg.Wait()
}

// Clear deletes all temporary files.
func (u *Uploads) Clear() {
	for _, f := range u.list {
		if f.TempFilename != "" && exists(f.TempFilename) {
			os.Remove(f.TempFilename)
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
func (f *FileUpload) Open(cfg *UploadsConfig) error {
	if cfg.Forbids(f.Name) {
		f.Error = UploadErrorExtension
		return nil
	}

	file, err := f.header.Open()
	if err != nil {
		f.Error = UploadErrorNoFile
		return err
	}
	defer file.Close()

	tmp, err := ioutil.TempFile(cfg.TmpDir(), "upload")
	if err != nil {
		// most likely cause of this issue is missing tmp dir
		f.Error = UploadErrorNoTmpDir
		return err
	}

	f.TempFilename = tmp.Name()
	defer tmp.Close()

	if f.Size, err = io.Copy(tmp, file); err != nil {
		f.Error = UploadErrorCantWrite
	}

	return err
}

// exists if file exists.
func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return false
}
