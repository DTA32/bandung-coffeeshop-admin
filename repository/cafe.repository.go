package repository

import (
	"context"
	"errors"

	"github.com/dta32/bandung-coffeeshop-admin/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CafeRepository owns the cafe aggregate: location (type cafe) + cafe + price,
// review, ratings, tags and images. Queries are ported from
// bandung-coffeeshop-be-console@252f8ed repository/admin_*.go.
type CafeRepository struct {
	db *pgxpool.Pool
}

func NewCafeRepository(db *pgxpool.Pool) *CafeRepository {
	return &CafeRepository{db: db}
}

func (r *CafeRepository) List(ctx context.Context, q string) ([]models.CafeListItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT l.id, c.id, l.name, l.status::text, c.is_featured, c.instagram,
		       TO_CHAR(GREATEST(l.updated_at, c.updated_at), 'YYYY-MM-DD')
		FROM location l
		JOIN cafe c ON c.location_id = l.id
		WHERE l.type = 'cafe'
		  AND ($1 = '' OR l.name ILIKE '%' || $1 || '%' OR l.id ILIKE '%' || $1 || '%')
		ORDER BY l.name ASC, l.id ASC
	`, q)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.CafeListItem, error) {
		var it models.CafeListItem
		err := row.Scan(&it.ID, &it.CafeID, &it.Name, &it.Status, &it.IsFeatured, &it.Instagram, &it.UpdatedAt)
		return it, err
	})
}

// Get assembles the full editable payload for a single cafe.
func (r *CafeRepository) Get(ctx context.Context, id string) (*models.Cafe, error) {
	var (
		p        models.Cafe
		lat, lng *float64
	)
	err := r.db.QueryRow(ctx, `
		SELECT l.id, l.name, l.description, l.description_indo, l.gmaps_id, l.status::text,
		       ST_Y(ST_Centroid(l.coordinates)), ST_X(ST_Centroid(l.coordinates)),
		       c.id, c.instagram, c.phone, c.menu,
		       TO_CHAR(c.open_hour, 'HH24:MI'), TO_CHAR(c.close_hour, 'HH24:MI'),
		       c.is_featured, c.weather
		FROM location l
		JOIN cafe c ON c.location_id = l.id
		WHERE l.id = $1 AND l.type = 'cafe'
	`, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.DescriptionIndo, &p.GmapsID, &p.Status,
		&lat, &lng,
		&p.CafeID, &p.Instagram, &p.Phone, &p.Menu, &p.OpenHour, &p.CloseHour,
		&p.IsFeatured, &p.Weather,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if lat != nil && lng != nil {
		p.Coordinates = &models.Coordinates{Lat: *lat, Lng: *lng}
	}

	err = r.db.QueryRow(ctx, `
		SELECT price_range_min, price_range_max, coffee_price_min, coffee_price_max,
		       snack_price_min, snack_price_max, food_price_min, food_price_max
		FROM cafe_price WHERE cafe_id = $1
	`, p.CafeID).Scan(
		&p.Price.PriceRangeMin, &p.Price.PriceRangeMax,
		&p.Price.CoffeePriceMin, &p.Price.CoffeePriceMax,
		&p.Price.SnackPriceMin, &p.Price.SnackPriceMax,
		&p.Price.FoodPriceMin, &p.Price.FoodPriceMax,
	)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	var rv models.Review
	err = r.db.QueryRow(ctx, `
		SELECT id, is_subjective, overall_score::float8, wfc_score::float8, personal_score::float8,
		       content, content_indo, visited_at::text
		FROM cafe_review
		WHERE cafe_id = $1
		ORDER BY COALESCE(visited_at, created_at::date) DESC, id DESC
		LIMIT 1
	`, p.CafeID).Scan(
		&rv.ID, &rv.IsSubjective, &rv.OverallScore, &rv.WFCScore,
		&rv.PersonalScore, &rv.Content, &rv.ContentIndo, &rv.VisitedAt,
	)
	if err == nil {
		p.Review = &rv
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `
		SELECT category_type, score::float8, short_description_override, short_description_override_indo
		FROM cafe_rating WHERE cafe_id = $1 ORDER BY category_type
	`, p.CafeID)
	if err != nil {
		return nil, err
	}
	p.Ratings, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Rating, error) {
		var rt models.Rating
		err := row.Scan(&rt.CategoryType, &rt.Score, &rt.ShortDescriptionOverride, &rt.ShortDescriptionOverrideIndo)
		return rt, err
	})
	if err != nil {
		return nil, err
	}

	rows, err = r.db.Query(ctx, `SELECT tag_id, visible FROM cafe_tag WHERE cafe_id = $1 ORDER BY tag_id`, p.CafeID)
	if err != nil {
		return nil, err
	}
	p.Tags, err = pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.CafeTag, error) {
		var tg models.CafeTag
		err := row.Scan(&tg.TagID, &tg.Visible)
		return tg, err
	})
	if err != nil {
		return nil, err
	}

	p.Images, err = listImages(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Create inserts a new cafe (location + cafe + children) in one transaction.
func (r *CafeRepository) Create(ctx context.Context, p models.Cafe) error {
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

	lng, lat := coordArgs(p.Coordinates)
	if _, err := tx.Exec(ctx, `
		INSERT INTO location (id, name, description, description_indo, gmaps_id, type, status, coordinates)
		VALUES ($1, $2, $3, $4, $5, 'cafe', $6::location_status_enum,
		        CASE WHEN $7::float8 IS NULL OR $8::float8 IS NULL THEN NULL
		             ELSE ST_SetSRID(ST_MakePoint($7, $8), 4326) END)
	`, p.ID, p.Name, p.Description, p.DescriptionIndo, p.GmapsID, p.Status, lng, lat); err != nil {
		return err
	}

	var cafeID int
	if err := tx.QueryRow(ctx, `
		INSERT INTO cafe (location_id, instagram, phone, menu, open_hour, close_hour, is_featured, weather)
		VALUES ($1, $2, $3, $4, $5::time, $6::time, $7, $8)
		RETURNING id
	`, p.ID, p.Instagram, p.Phone, p.Menu, p.OpenHour, p.CloseHour, p.IsFeatured, p.Weather).Scan(&cafeID); err != nil {
		return err
	}

	if err := writeCafeChildren(ctx, tx, p.ID, cafeID, p); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// Update rewrites every table for an existing cafe in one transaction.
func (r *CafeRepository) Update(ctx context.Context, id string, p models.Cafe) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var cafeID int
	if err := tx.QueryRow(ctx, `
		SELECT c.id FROM cafe c JOIN location l ON l.id = c.location_id
		WHERE l.id = $1 AND l.type = 'cafe'
	`, id).Scan(&cafeID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	lng, lat := coordArgs(p.Coordinates)
	if _, err := tx.Exec(ctx, `
		UPDATE location SET
			name = $2, description = $3, description_indo = $4, gmaps_id = $5,
			status = $6::location_status_enum,
			coordinates = CASE WHEN $7::float8 IS NULL OR $8::float8 IS NULL THEN NULL
			                   ELSE ST_SetSRID(ST_MakePoint($7, $8), 4326) END,
			updated_at = NOW()
		WHERE id = $1 AND type = 'cafe'
	`, id, p.Name, p.Description, p.DescriptionIndo, p.GmapsID, p.Status, lng, lat); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		UPDATE cafe SET
			instagram = $2, phone = $3, menu = $4, open_hour = $5::time, close_hour = $6::time,
			is_featured = $7, weather = $8, updated_at = NOW()
		WHERE id = $1
	`, cafeID, p.Instagram, p.Phone, p.Menu, p.OpenHour, p.CloseHour, p.IsFeatured, p.Weather); err != nil {
		return err
	}

	if err := writeCafeChildren(ctx, tx, id, cafeID, p); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SoftDelete marks the cafe's location as deleted.
func (r *CafeRepository) SoftDelete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE location SET status = 'deleted', updated_at = NOW()
		WHERE id = $1 AND type = 'cafe'
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func coordArgs(c *models.Coordinates) (*float64, *float64) {
	if c == nil {
		return nil, nil
	}
	lng, lat := c.Lng, c.Lat
	return &lng, &lat
}

// writeCafeChildren upserts price/review/ratings/tags/images for a cafe,
// deleting ratings, tags and images absent from the payload.
func writeCafeChildren(ctx context.Context, tx pgx.Tx, locationID string, cafeID int, p models.Cafe) error {
	if _, err := tx.Exec(ctx, `
		INSERT INTO cafe_price (cafe_id, price_range_min, price_range_max,
			coffee_price_min, coffee_price_max, snack_price_min, snack_price_max,
			food_price_min, food_price_max)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (cafe_id) DO UPDATE SET
			price_range_min = EXCLUDED.price_range_min,
			price_range_max = EXCLUDED.price_range_max,
			coffee_price_min = EXCLUDED.coffee_price_min,
			coffee_price_max = EXCLUDED.coffee_price_max,
			snack_price_min = EXCLUDED.snack_price_min,
			snack_price_max = EXCLUDED.snack_price_max,
			food_price_min = EXCLUDED.food_price_min,
			food_price_max = EXCLUDED.food_price_max,
			updated_at = NOW()
	`, cafeID,
		p.Price.PriceRangeMin, p.Price.PriceRangeMax,
		p.Price.CoffeePriceMin, p.Price.CoffeePriceMax,
		p.Price.SnackPriceMin, p.Price.SnackPriceMax,
		p.Price.FoodPriceMin, p.Price.FoodPriceMax,
	); err != nil {
		return err
	}

	if rv := p.Review; rv != nil {
		if rv.ID != nil {
			if _, err := tx.Exec(ctx, `
				UPDATE cafe_review SET
					is_subjective = $2, overall_score = $3, wfc_score = $4,
					personal_score = $5, content = $6, content_indo = $7,
					visited_at = $8::date, updated_at = NOW()
				WHERE id = $1 AND cafe_id = $9
			`, *rv.ID, rv.IsSubjective, rv.OverallScore, rv.WFCScore,
				rv.PersonalScore, rv.Content, rv.ContentIndo, rv.VisitedAt, cafeID); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec(ctx, `
				INSERT INTO cafe_review (cafe_id, is_subjective, overall_score,
					wfc_score, personal_score, content, content_indo, visited_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8::date)
			`, cafeID, rv.IsSubjective, rv.OverallScore, rv.WFCScore,
				rv.PersonalScore, rv.Content, rv.ContentIndo, rv.VisitedAt); err != nil {
				return err
			}
		}
	}

	types := make([]string, 0, len(p.Ratings))
	for _, rt := range p.Ratings {
		types = append(types, rt.CategoryType)
		if _, err := tx.Exec(ctx, `
			INSERT INTO cafe_rating (cafe_id, category_type, score,
				short_description_override, short_description_override_indo)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (cafe_id, category_type) DO UPDATE SET
				score = EXCLUDED.score,
				short_description_override = EXCLUDED.short_description_override,
				short_description_override_indo = EXCLUDED.short_description_override_indo,
				updated_at = NOW()
		`, cafeID, rt.CategoryType, rt.Score, rt.ShortDescriptionOverride, rt.ShortDescriptionOverrideIndo); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM cafe_rating WHERE cafe_id = $1 AND category_type <> ALL($2::text[])
	`, cafeID, types); err != nil {
		return err
	}

	tagIDs := make([]int, 0, len(p.Tags))
	for _, tg := range p.Tags {
		tagIDs = append(tagIDs, tg.TagID)
		if _, err := tx.Exec(ctx, `
			INSERT INTO cafe_tag (cafe_id, tag_id, visible)
			VALUES ($1, $2, $3)
			ON CONFLICT (cafe_id, tag_id) DO UPDATE SET
				visible = EXCLUDED.visible,
				updated_at = CASE WHEN cafe_tag.visible IS DISTINCT FROM EXCLUDED.visible
				                  THEN NOW() ELSE cafe_tag.updated_at END
		`, cafeID, tg.TagID, tg.Visible); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `
		DELETE FROM cafe_tag WHERE cafe_id = $1 AND tag_id <> ALL($2::int[])
	`, cafeID, tagIDs); err != nil {
		return err
	}

	return writeLocationImages(ctx, tx, locationID, p.Images)
}

// IDByLocation resolves a cafe's numeric id from its location id.
func (r *CafeRepository) IDByLocation(ctx context.Context, locationID string) (int, string, error) {
	var id int
	var name string
	err := r.db.QueryRow(ctx, `
		SELECT c.id, l.name FROM cafe c JOIN location l ON l.id = c.location_id WHERE l.id = $1
	`, locationID).Scan(&id, &name)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, "", ErrNotFound
	}
	return id, name, err
}
