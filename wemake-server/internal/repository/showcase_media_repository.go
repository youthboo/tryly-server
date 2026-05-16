package repository

import (
	"database/sql"
	"strings"

	"github.com/yourusername/wemake/internal/domain"
)

func (r *ShowcaseRepository) CreateImage(img *domain.ShowcaseImage, factoryID int64) error {
	var head struct {
		FactoryID       int64                `db:"factory_id"`
		LinkedShowcases domain.JSONLinkArray `db:"linked_showcases"`
	}
	if err := r.db.Get(&head, `
		SELECT factory_id, linked_showcases
		FROM factory_showcases
		WHERE showcase_id = $1
	`, img.ShowcaseID); err != nil {
		return sql.ErrNoRows
	}
	if head.FactoryID != factoryID {
		return domain.ErrForbidden
	}

	imageURL := strings.TrimSpace(img.ImageURL)
	if imageURL == "" {
		return nil
	}

	linked := make([]string, 0, len(head.LinkedShowcases)+1)
	seen := map[string]struct{}{}
	for _, raw := range head.LinkedShowcases {
		v := strings.TrimSpace(raw)
		if v == "" || v == imageURL {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		linked = append(linked, v)
	}

	insertAt := img.SortOrder
	if insertAt <= 0 {
		insertAt = len(linked)
	} else {
		insertAt--
	}
	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(linked) {
		insertAt = len(linked)
	}
	linked = append(linked, "")
	copy(linked[insertAt+1:], linked[insertAt:])
	linked[insertAt] = imageURL

	if len(linked) > 5 {
		return domain.ErrImageLimitExceeded
	}

	res, err := r.db.Exec(`
		UPDATE factory_showcases
		SET linked_showcases = $1,
		    updated_at = NOW()
		WHERE showcase_id = $2 AND factory_id = $3
	`, domain.JSONLinkArray(linked), img.ShowcaseID, factoryID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	img.ImageURL = imageURL
	return nil
}

func (r *ShowcaseRepository) ListImages(showcaseID, callerID int64) ([]domain.ShowcaseImage, error) {
	var head struct {
		FactoryID       int64                `db:"factory_id"`
		Status          string               `db:"status"`
		LinkedShowcases domain.JSONLinkArray `db:"linked_showcases"`
	}
	if err := r.db.Get(&head, `
		SELECT factory_id, status, linked_showcases
		FROM factory_showcases
		WHERE showcase_id = $1
	`, showcaseID); err != nil {
		return nil, err
	}
	if head.Status != "AC" && callerID != head.FactoryID {
		return nil, sql.ErrNoRows
	}

	images := []domain.ShowcaseImage{}
	sortOrder := 0
	seen := map[string]struct{}{}
	addVirtual := func(raw string) {
		url := strings.TrimSpace(raw)
		if url == "" {
			return
		}
		lower := strings.ToLower(url)
		if !strings.HasPrefix(lower, "https://") && !strings.HasPrefix(lower, "http://") {
			return
		}
		if _, ok := seen[url]; ok {
			return
		}
		seen[url] = struct{}{}
		images = append(images, domain.ShowcaseImage{
			ShowcaseID: showcaseID,
			ImageURL:   url,
			SortOrder:  sortOrder,
		})
		sortOrder++
	}
	for _, ref := range head.LinkedShowcases {
		addVirtual(ref)
	}
	return images, nil
}

func (r *ShowcaseRepository) DeleteImage(showcaseID, imageID, factoryID int64) error {
	res, err := r.db.Exec(`
		DELETE FROM showcase_images
		WHERE image_id = $1
		  AND showcase_id = $2
		  AND EXISTS (SELECT 1 FROM factory_showcases WHERE showcase_id = $2 AND factory_id = $3)
	`, imageID, showcaseID, factoryID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *ShowcaseRepository) PatchImage(showcaseID, imageID, factoryID int64, sortOrder *int, caption *string) (*domain.ShowcaseImage, error) {
	res, err := r.db.Exec(`
		UPDATE showcase_images
		SET sort_order = COALESCE($1, sort_order),
		    caption    = COALESCE($2, caption)
		WHERE image_id = $3
		  AND showcase_id = $4
		  AND EXISTS (SELECT 1 FROM factory_showcases WHERE showcase_id = $4 AND factory_id = $5)
	`, sortOrder, caption, imageID, showcaseID, factoryID)
	if err != nil {
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, sql.ErrNoRows
	}
	var img domain.ShowcaseImage
	err = r.db.Get(&img, `
		SELECT image_id, showcase_id, image_url, sort_order, caption
		FROM showcase_images WHERE image_id = $1
	`, imageID)
	return &img, err
}
