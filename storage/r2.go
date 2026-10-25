// Package storage uploads images to Cloudflare R2 through its S3 API (the same
// bucket the scraper writes ig/ media to). Only a conditional PUT is needed, so
// requests are signed here with AWS SigV4 instead of pulling in an S3 SDK.
package storage

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

// ErrExists is returned by PutNew when the key is already taken.
var ErrExists = errors.New("object already exists")

type R2 struct {
	endpoint  string // scheme + host, no trailing slash
	region    string
	bucket    string
	accessKey string
	secretKey string
	client    *http.Client
}

func NewR2(accountID, accessKeyID, secretAccessKey, bucket string) *R2 {
	return &R2{
		endpoint:  "https://" + accountID + ".r2.cloudflarestorage.com",
		region:    "auto",
		bucket:    bucket,
		accessKey: accessKeyID,
		secretKey: secretAccessKey,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

// PutNew uploads body under key unless an object already exists there
// (If-None-Match: *), so an upload never replaces an image already in use.
// key must be URL-path safe (letters, digits, '.', '_', '-', '/').
func (s *R2) PutNew(ctx context.Context, key, contentType string, body []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut,
		s.endpoint+"/"+s.bucket+"/"+key, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("If-None-Match", "*")
	sum := sha256.Sum256(body)
	s.sign(req, hex.EncodeToString(sum[:]), time.Now().UTC())

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("r2 put %s: %w", key, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusPreconditionFailed:
		return ErrExists
	case resp.StatusCode/100 != 2:
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("r2 put %s: %s: %s", key, resp.Status, bytes.TrimSpace(msg))
	}
	return nil
}

// sign adds the SigV4 headers (x-amz-date, x-amz-content-sha256,
// Authorization), signing host plus every header already set on req.
func (s *R2) sign(req *http.Request, payloadHash string, now time.Time) {
	amzDate := now.Format("20060102T150405Z")
	date := now.Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)

	headers := map[string]string{"host": req.URL.Host}
	for k, v := range req.Header {
		headers[strings.ToLower(k)] = strings.TrimSpace(strings.Join(v, ","))
	}
	names := make([]string, 0, len(headers))
	for k := range headers {
		names = append(names, k)
	}
	sort.Strings(names)
	var canonHeaders strings.Builder
	for _, k := range names {
		canonHeaders.WriteString(k + ":" + headers[k] + "\n")
	}
	signed := strings.Join(names, ";")

	canonReq := strings.Join([]string{
		req.Method,
		req.URL.EscapedPath(),
		req.URL.RawQuery,
		canonHeaders.String(),
		signed,
		payloadHash,
	}, "\n")
	scope := date + "/" + s.region + "/s3/aws4_request"
	reqHash := sha256.Sum256([]byte(canonReq))
	toSign := "AWS4-HMAC-SHA256\n" + amzDate + "\n" + scope + "\n" + hex.EncodeToString(reqHash[:])

	k := hmacSHA256([]byte("AWS4"+s.secretKey), date)
	k = hmacSHA256(k, s.region)
	k = hmacSHA256(k, "s3")
	k = hmacSHA256(k, "aws4_request")
	sig := hex.EncodeToString(hmacSHA256(k, toSign))

	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+s.accessKey+"/"+scope+
		",SignedHeaders="+signed+",Signature="+sig)
}

func hmacSHA256(key []byte, data string) []byte {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(data))
	return m.Sum(nil)
}
