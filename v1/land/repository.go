package land

import (
	"context"
	"math"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// ILandRepository Details the methods required for the Land repository
type ILandRepository interface {
	FindDynamicTile(ctx context.Context, tile tileInput) ([]byte, error)
}

// Repository allowing for dependency injection
type Repository struct {
	Db *sqlx.DB
}

// FindLandParcel Returns a single land feature based on Title Number
func (f *Repository) FindDynamicTile(ctx context.Context, tile tileInput) ([]byte, error) {

	e := tileToEnv(tile)

	segSize := (e.xMax - e.xMin) / 4

	rows, err := f.Db.QueryxContext(ctx, `
	WITH
            bounds AS (
                SELECT ST_Segmentize(ST_MakeEnvelope($1, $2, $3, $4, 3857), $5) AS geom,
                       ST_Segmentize(ST_MakeEnvelope($1, $2, $3, $4, 3857), $5)::box2d AS b2d
            ),
            mvtgeom AS (
                SELECT ST_AsMVTGeom(ST_Transform(t.geometry, 3857), bounds.b2d) AS geom,
                       uprns, tenure, height, planning_use_classes, property_address,ownership, proprietor_ids,is_foreign_co, area_sqft, area_buildings_sqft, density, risk_level, has_defence, is_listed
                FROM land_v2 t, bounds
                WHERE ST_Intersects(t.geometry, ST_Transform(bounds.geom, 4326))
                AND t.area_sqft > 43560.000000
                AND t.area_sqft < 130680.000000
                AND t.density > 0.000000
                AND t.density < 100.000000
            )
            SELECT ST_AsMVT(mvtgeom.*) FROM mvtgeom`, e.xMin, e.yMin, e.xMax, e.yMax, segSize)
	if err != nil {
		return nil, errors.Wrap(err, "FindDynamicTile: Query failed")
	}
	defer rows.Close()

	var mvt []byte
	for rows.Next() {

		err = rows.Scan(
			&mvt,
		)
		if err != nil {
			return nil, errors.Wrap(err, "FindDynamicTile: Scan failed")
		}
	}

	return mvt, nil
}

func tileToEnv(tile tileInput) env {
	var e env

	worldMercMax := 20037508.3427892
	worldMercMin := -1 * worldMercMax
	worldMercSize := worldMercMax - worldMercMin

	worldTileSize := math.Pow(2, float64(*tile.zoom))

	tileMercSize := worldMercSize / worldTileSize

	e.xMin = worldMercMin + tileMercSize*float64(*tile.x)
	e.xMax = worldMercMin + tileMercSize*float64(*tile.x+1)
	e.yMin = worldMercMax - tileMercSize*float64(*tile.y+1)
	e.yMax = worldMercMax - tileMercSize*float64(*tile.y)
	return e
}
