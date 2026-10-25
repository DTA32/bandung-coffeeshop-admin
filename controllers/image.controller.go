package controllers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/storage"
)

// MaxImageBytes caps one uploaded image.
const MaxImageBytes = 15 << 20

// ImageFolders are the R2 prefixes an upload may target: <folder>/<slug>/<file>.
var ImageFolders = []string{"cafe", "location"}

// imageTypes maps the sniffed content type to the extensions kept for it (the
// first one is used when the file name has none of them).
var imageTypes = map[string][]string{
	"image/jpeg": {".jpg", ".jpeg"},
	"image/png":  {".png"},
	"image/webp": {".webp"},
}

// ImageController uploads cafe / location images to R2. store is nil when R2
// is not configured; the forms then only take typed URLs.
type ImageController struct {
	store *storage.R2
	base  string // public image host, e.g. https://image.bdgcafe.com
}

func NewImageController(store *storage.R2, base string) *ImageController {
	return &ImageController{store: store, base: base}
}

func (c *ImageController) Enabled() bool { return c.store != nil }

// Upload stores body at <folder>/<slug>/<file name> and returns its public URL.
// The original file name is kept (sanitised); a taken name gets a -2, -3...
// suffix instead of replacing the existing object.
func (c *ImageController) Upload(ctx context.Context, folder, slug, filename string, body []byte) (string, error) {
	if c.store == nil {
		return "", invalid("image uploads are not configured")
	}
	if !slices.Contains(ImageFolders, folder) {
		return "", invalid("invalid image folder %q", folder)
	}
	if !slugRe.MatchString(slug) {
		return "", invalid("set a valid ID / slug before uploading images")
	}
	if len(body) == 0 {
		return "", invalid("the file is empty")
	}
	if len(body) > MaxImageBytes {
		return "", invalid("images must be at most %d MB", MaxImageBytes>>20)
	}
	ct := http.DetectContentType(body)
	exts, ok := imageTypes[ct]
	if !ok {
		return "", invalid("only JPEG, PNG and WebP images can be uploaded (got %s)", ct)
	}
	name := ImageFileName(filename, exts)
	prefix := folder + "/" + slug + "/"
	for i := 1; i <= 50; i++ {
		key := prefix + withSuffix(name, i)
		err := c.store.PutNew(ctx, key, ct, body)
		if errors.Is(err, storage.ErrExists) {
			continue
		}
		if err != nil {
			return "", err
		}
		return c.base + "/" + key, nil
	}
	return "", invalid("too many images named %s, rename the file", name)
}

var unsafeNameRe = regexp.MustCompile(`[^A-Za-z0-9._-]+`)

// ImageFileName turns an uploaded file name into a URL-safe key segment:
// runs of other characters become '-', and the extension is one of exts
// (kept when the name already has one, else exts[0]).
func ImageFileName(filename string, exts []string) string {
	base := path.Base(strings.ReplaceAll(filename, `\`, "/"))
	ext := strings.ToLower(path.Ext(base))
	stem := strings.TrimSuffix(base, path.Ext(base))
	if !slices.Contains(exts, ext) {
		ext = exts[0]
	}
	stem = strings.Trim(unsafeNameRe.ReplaceAllString(stem, "-"), "-._")
	if stem == "" {
		stem = "image"
	}
	return stem + ext
}

// withSuffix returns name for i == 1, else name with -i before the extension.
func withSuffix(name string, i int) string {
	if i == 1 {
		return name
	}
	ext := path.Ext(name)
	return fmt.Sprintf("%s-%d%s", strings.TrimSuffix(name, ext), i, ext)
}
