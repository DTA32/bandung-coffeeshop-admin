package controllers

import (
	"testing"

	"github.com/dta32/bandung-coffeeshop-admin/models"
)

func TestNormalizeLocation(t *testing.T) {
	l := models.Location{Name: "Gedung Sate", Type: "poi", Geojson: ptr(`{"type":"Point","coordinates":[107.6,-6.9]}`)}
	if err := NormalizeLocation(&l); err != nil {
		t.Fatal(err)
	}
	l = models.Location{Name: "Area", Type: "area", Geojson: ptr(`{"type":"Point","coordinates":[1,2]}`)}
	if err := NormalizeLocation(&l); err == nil {
		t.Fatal("area with a Point must be rejected")
	}
	l = models.Location{Name: "X", Type: "cafe"}
	if err := NormalizeLocation(&l); err == nil {
		t.Fatal("cafes are not edited as locations")
	}
	l = models.Location{Name: "X", Type: "poi", Geojson: ptr(`not json`)}
	if err := NormalizeLocation(&l); err == nil {
		t.Fatal("invalid json must be rejected")
	}
}
