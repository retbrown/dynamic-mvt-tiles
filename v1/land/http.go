package land

import (
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// HTTP allowing for dependency injection
type HTTP struct {
	Repository ILandRepository
}

func (land *HTTP) GetTile(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	var tile tileInput

	zString := chi.URLParam(r, "z")
	z, err := strconv.Atoi(zString)
	if err != nil {
		logger.Error("GetLandParcel: z error", zap.Error(err))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tile.zoom = &z

	xString := chi.URLParam(r, "x")
	x, err := strconv.Atoi(xString)
	if err != nil {
		logger.Error("GetLandParcel: x error", zap.Error(err))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tile.x = &x

	yString := chi.URLParam(r, "y")
	y, err := strconv.Atoi(yString)
	if err != nil {
		logger.Error("GetLandParcel: y error", zap.Error(err))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	tile.y = &y

	extension := chi.URLParam(r, "extension")
	tile.format = &extension

	if !tileIsValid(logger, tile) {
		logger.Error("GetLandParcel: invalid tile")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Info("Request", zap.String("path", r.URL.Path), zap.Any("tile", tile))

	response, err := land.Repository.FindDynamicTile(r.Context(), tile)
	if err != nil {
		logger.Error("GetLandParcel: Service error", zap.Error(err))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write bytes back to the caller
	w.Header().Set("Content-Type", "application/vnd.mapbox-vector-tile")
	w.Write(response)
	return
}

func tileIsValid(logger *zap.Logger, tile tileInput) bool {
	if tile.x == nil || tile.y == nil || tile.zoom == nil || tile.format == nil {
		logger.Error("Item missing from tile")
		return false
	}

	if *tile.format != "pbf" && *tile.format != "mvt" {
		logger.Error("Format wrong", zap.String("read format", *tile.format))
		return false
	}

	size := math.Pow(2, float64(*tile.zoom))

	if float64(*tile.x) >= size || float64(*tile.y) >= size {
		logger.Error("X or Y too large")
		return false
	}

	if *tile.x < 0 || *tile.y < 0 {
		logger.Error("X or Y less than 0")
		return false
	}

	return true
}
