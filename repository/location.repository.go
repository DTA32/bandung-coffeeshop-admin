package repository

import (
	"context"
	"errors"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LocationRepository owns non-cafe locations (poi / area / district) and the
// location_image rows shared with cafes.
type LocationRepository struct {
	db *pgxpool.Pool
}

func NewLocationRepository(db *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{db: db}
}

// List returns one page (1-based) of non-cafe locations, optionally of one
// type, whose name or id contains q (case-insensitive, wildcards escaped).
func (r *LocationRepository) List(ctx context.Context, typ, q string, page int) (models.Page[models.LocationListItem], error) {
	limit, offset := pageWindow(page)
	rows, err := r.db.Query(ctx, `
		SELECT id, name, type::text, status::text
		FROM location
		WHERE type <> 'cafe' AND ($1 = '' OR type::text = $1)
		  AND ($2::text = '' OR name ILIKE '%' || $2 || '%' OR id ILIKE '%' || $2 || '%')
		ORDER BY type, name ASC, id ASC
		LIMIT $3 OFFSET $4
	`, typ, EscapeLike(q), limit, offset)
	if err != nil {
		return models.Page[models.LocationListItem]{}, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.LocationListItem, error) {
		var it models.LocationListItem
		err := row.Scan(&it.ID, &it.Name, &it.Type, &it.Status)
		return it, err
	})
	return toPage(items), err
}

func (r *LocationRepository) Get(ctx context.Context, id string) (*models.Location, error) {
	var p models.Location
	err := r.db.QueryRow(ctx, `
		SELECT id, name, description, description_indo, gmaps_id, type::text, status::text,
		       ST_AsGeoJSON(coordinates)
		FROM location
		WHERE id = $1 AND type <> 'cafe'
	`, id).Scan(&p.ID, &p.Name, &p.Description, &p.DescriptionIndo, &p.GmapsID, &p.Type, &p.Status, &p.Geojson)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	p.Images, err = listImages(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *LocationRepository) Create(ctx context.Context, p models.Location) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM location WHERE id = $1)`, p.ID).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrIDExists
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO location (id, name, description, description_indo, gmaps_id, type, status, coordinates)
		VALUES ($1, $2, $3, $4, $5, $6::location_type_enum, $7::location_status_enum,
		        CASE WHEN $8::text IS NULL OR $8::text = '' THEN NULL
		             ELSE ST_SetSRID(ST_GeomFromGeoJSON($8), 4326) END)
	`, p.ID, p.Name, p.Description, p.DescriptionIndo, p.GmapsID, p.Type, p.Status, p.Geojson); err != nil {
		return err
	}
	if err := writeLocationImages(ctx, tx, p.ID, p.Images); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LocationRepository) Update(ctx context.Context, id string, p models.Location) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `
		UPDATE location SET
			name = $2, description = $3, description_indo = $4, gmaps_id = $5,
			type = $6::location_type_enum, status = $7::location_status_enum,
			coordinates = CASE WHEN $8::text IS NULL OR $8::text = '' THEN NULL
			                   ELSE ST_SetSRID(ST_GeomFromGeoJSON($8), 4326) END,
			updated_at = NOW()
		WHERE id = $1 AND type <> 'cafe'
	`, id, p.Name, p.Description, p.DescriptionIndo, p.GmapsID, p.Type, p.Status, p.Geojson)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	if err := writeLocationImages(ctx, tx, id, p.Images); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *LocationRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE location SET status = 'deleted', updated_at = NOW()
		WHERE id = $1 AND type <> 'cafe'
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// listImages returns a location's images in display order (shared with cafes).
func listImages(ctx context.Context, db *pgxpool.Pool, locationID string) ([]models.Image, error) {
	rows, err := db.Query(ctx, `
		SELECT url, description, description_indo
		FROM location_image WHERE location_id = $1
		ORDER BY display_order ASC, id ASC
	`, locationID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Image, error) {
		var im models.Image
		err := row.Scan(&im.URL, &im.Description, &im.DescriptionIndo)
		return im, err
	})
}

// writeLocationImages upserts images keyed by (location_id, url) and deletes
// any image absent from the payload; display_order follows slice order.
func writeLocationImages(ctx context.Context, tx pgx.Tx, locationID string, images []models.Image) error {
	urls := make([]string, 0, len(images))
	for i, im := range images {
		urls = append(urls, im.URL)
		if _, err := tx.Exec(ctx, `
			INSERT INTO location_image (location_id, url, description, description_indo, display_order)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (location_id, url) DO UPDATE SET
				description = EXCLUDED.description,
				description_indo = EXCLUDED.description_indo,
				display_order = EXCLUDED.display_order,
				updated_at = NOW()
		`, locationID, im.URL, im.Description, im.DescriptionIndo, i); err != nil {
			return err
		}
	}
	_, err := tx.Exec(ctx, `
		DELETE FROM location_image WHERE location_id = $1 AND url <> ALL($2::text[])
	`, locationID, urls)
	return err
}
