package static

import (
	"hash/crc32"
	"io"
	"net/http"
	"os"
	"unsafe"

	httpConfig "github.com/spiral/roadrunner/v2/plugins/http/config"
)

const etag string = "Etag"

// weak Etag prefix
var weakPrefix = []byte(`W/`)

// CRC32 table
var crc32q = crc32.MakeTable(0x48D90782)

func SetEtag(cfg *httpConfig.Static, f *os.File, w http.ResponseWriter) {
	// read the file content
	body, err := io.ReadAll(f)
	if err != nil {
		return
	}

	// skip for 0 body
	if len(body) == 0 {
		return
	}

	// preallocate
	calculatedEtag := make([]byte, 0, 64)

	// write weak
	if cfg.Weak {
		calculatedEtag = append(calculatedEtag, weakPrefix...)
	}

	calculatedEtag = append(calculatedEtag, '"')
	calculatedEtag = appendUint(calculatedEtag, uint32(len(body)))
	calculatedEtag = append(calculatedEtag, '-')
	calculatedEtag = appendUint(calculatedEtag, crc32.Checksum(body, crc32q))
	calculatedEtag = append(calculatedEtag, '"')

	w.Header().Set(etag, byteToSrt(calculatedEtag))
}

// appendUint appends n to dst and returns the extended dst.
func appendUint(dst []byte, n uint32) []byte {
	var b [20]byte
	buf := b[:]
	i := len(buf)
	var q uint32
	for n >= 10 {
		i--
		q = n / 10
		buf[i] = '0' + byte(n-q*10)
		n = q
	}
	i--
	buf[i] = '0' + byte(n)

	dst = append(dst, buf[i:]...)
	return dst
}

func byteToSrt(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
