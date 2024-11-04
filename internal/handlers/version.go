// internal/handlers/version_handler.go
package handlers

import (
	"cor-events-scheduler/internal/domain/models"
	"cor-events-scheduler/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type VersionHandler struct {
	versionService *services.VersionService
	logger         *zap.Logger
}

func NewVersionHandler(versionService *services.VersionService, logger *zap.Logger) *VersionHandler {
	return &VersionHandler{
		versionService: versionService,
		logger:         logger,
	}
}

// @Summary Get schedule version history
// @Description Get the version history of a schedule
// @Tags versions
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Success 200 {array} models.VersionMetadata
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /schedules/{id}/versions [get]
func (h *VersionHandler) GetVersionHistory(c *gin.Context) {
	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid schedule ID format",
			zap.Error(err),
			zap.String("schedule_id", c.Param("id")),
		)
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid schedule ID format",
		})
		return
	}

	versions, err := h.versionService.GetVersionHistory(c.Request.Context(), uint(scheduleID))
	if err != nil {
		h.logger.Error("Failed to get version history",
			zap.Error(err),
			zap.Uint64("schedule_id", scheduleID),
		)

		// Check if it's a not found error
		if err.Error() == "record not found" {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Schedule not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to get version history",
			Details: err.Error(),
		})
		return
	}

	// If no versions found, return empty array instead of null
	if versions == nil {
		versions = []models.VersionMetadata{}
	}

	h.logger.Debug("Retrieved version history",
		zap.Uint64("schedule_id", scheduleID),
		zap.Int("version_count", len(versions)),
	)

	c.JSON(http.StatusOK, versions)
}

// @Summary Restore schedule version
// @Description Restore a specific version of a schedule
// @Tags versions
// @Accept json
// @Produce json
// @Param id path int true "Schedule ID"
// @Param version path int true "Version number"
// @Success 200 {object} models.Schedule
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /schedules/{id}/versions/{version}/restore [post]
func (h *VersionHandler) RestoreVersion(c *gin.Context) {
	scheduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		h.logger.Error("Invalid schedule ID format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid schedule ID format",
		})
		return
	}

	version, err := strconv.Atoi(c.Param("version"))
	if err != nil {
		h.logger.Error("Invalid version format", zap.Error(err))
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid version format",
		})
		return
	}

	err = h.versionService.RestoreVersion(c.Request.Context(), uint(scheduleID), version)
	if err != nil {
		h.logger.Error("Failed to restore version",
			zap.Error(err),
			zap.Uint64("schedule_id", scheduleID),
			zap.Int("version", version),
		)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to restore version",
			Details: err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}
