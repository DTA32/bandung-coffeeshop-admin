package tag

import (
	"net/url"
	"strings"

	"github.com/dta32/bandung-coffeeshop-admin/handlers"
	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func ParseTagForm(v url.Values) models.Tag {
	return models.Tag{
		Name:            v.Get("name"),
		NameIndo:        strings.TrimSpace(v.Get("name_indo")),
		Slug:            handlers.OptString(v, "slug"),
		Description:     strings.TrimSpace(v.Get("description")),
		DescriptionIndo: strings.TrimSpace(v.Get("description_indo")),
	}
}
