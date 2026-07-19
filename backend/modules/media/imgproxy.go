package media

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"

	"github.com/industrix/backend/pkg/errors"
)

// Presets are the named image variants callers can ask for. Keeping the set
// closed (rather than accepting arbitrary width/height) stops the proxy from
// being used to generate unbounded renditions.
var presets = map[string]string{
	"thumb": "rs:fill:300:300:0/g:sm", // square crop for grids/lists
	"card":  "rs:fit:600:600:0",       // listing card
	"full":  "rs:fit:1200:1200:0",     // detail view
}

// Imgproxy builds signed imgproxy URLs for images stored in our bucket.
//
// Two hosts are involved and they are not interchangeable: the stored/public
// URL is the one the BROWSER uses (localhost:9000), while imgproxy fetches the
// source itself and must use the internal docker host (minio:9000). So the
// source is rewritten from the public prefix to the internal one before it is
// signed.
type Imgproxy struct {
	baseURL        string // imgproxy itself, browser-reachable (e.g. http://localhost:8082)
	key            []byte // hex-decoded IMGPROXY_KEY
	salt           []byte // hex-decoded IMGPROXY_SALT
	publicPrefix   string // http://localhost:9000/equipment-media
	internalPrefix string // http://minio:9000/equipment-media
}

// NewImgproxy builds the URL signer. key/salt are hex strings; when either is
// empty the "insecure" signature is used (imgproxy must then run with
// signature checking disabled).
func NewImgproxy(baseURL, keyHex, saltHex, publicPrefix, internalPrefix string) *Imgproxy {
	key, _ := hex.DecodeString(keyHex)
	salt, _ := hex.DecodeString(saltHex)
	return &Imgproxy{
		baseURL:        strings.TrimRight(baseURL, "/"),
		key:            key,
		salt:           salt,
		publicPrefix:   strings.TrimRight(publicPrefix, "/"),
		internalPrefix: strings.TrimRight(internalPrefix, "/"),
	}
}

// Enabled reports whether a proxy base URL was configured.
func (p *Imgproxy) Enabled() bool { return p != nil && p.baseURL != "" }

// URL returns the signed imgproxy URL for a source image and preset.
// The source must live in our own bucket — this is the guard that keeps
// imgproxy from being turned into an open proxy for arbitrary URLs.
func (p *Imgproxy) URL(src, preset string) (string, error) {
	if !p.Enabled() {
		return "", errors.New(errors.CodeInternal, "Image processing is not configured")
	}
	opts, ok := presets[preset]
	if !ok {
		return "", errors.New(errors.CodeValidation, "Unknown preset — use thumb, card or full")
	}

	// Accept either form of our own URL, and always hand imgproxy the internal one.
	var source string
	switch {
	case strings.HasPrefix(src, p.publicPrefix):
		source = p.internalPrefix + strings.TrimPrefix(src, p.publicPrefix)
	case strings.HasPrefix(src, p.internalPrefix):
		source = src
	default:
		return "", errors.New(errors.CodeValidation, "Image is not from this service")
	}

	encoded := base64.RawURLEncoding.EncodeToString([]byte(source))
	path := "/" + opts + "/" + encoded

	signature := "insecure"
	if len(p.key) > 0 && len(p.salt) > 0 {
		mac := hmac.New(sha256.New, p.key)
		mac.Write(p.salt)
		mac.Write([]byte(path))
		signature = base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	}
	return p.baseURL + "/" + signature + path, nil
}
