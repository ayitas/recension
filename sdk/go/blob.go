package recension

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/ayitas/recension/pkg/message"
)

const digestPrefix = "sha256:"

// BlobFile is binary content with an optional MIME type for Check/Assume.
type BlobFile struct {
	Data []byte
	Mime string
}

func digestOf(data []byte) string {
	sum := sha256.Sum256(data)
	return digestPrefix + hex.EncodeToString(sum[:])
}

func (c *client) resolveValue(value any) (message.Value, error) {
	switch x := value.(type) {
	case []byte:
		return c.ensureBlob(x, "application/octet-stream")
	case BlobFile:
		mime := x.Mime
		if mime == "" {
			mime = "application/octet-stream"
		}
		return c.ensureBlob(x.Data, mime)
	default:
		return toValue(value), nil
	}
}

func (c *client) ensureBlob(data []byte, mime string) (message.Value, error) {
	dig := digestOf(data)
	if c.opts.Offline {
		return message.BlobValue(dig, mime), nil
	}
	path := "/v1/blobs/" + url.PathEscape(dig)
	exists, err := c.transport.head(path)
	if err != nil {
		return message.Value{}, fmt.Errorf("recension: blob head: %w", err)
	}
	if !exists {
		if err := c.transport.putRaw(path, data, mime); err != nil {
			return message.Value{}, fmt.Errorf("recension: blob upload: %w", err)
		}
	}
	return message.BlobValue(dig, mime), nil
}
