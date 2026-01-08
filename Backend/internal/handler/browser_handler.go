package handler

import (
	"gbackup-new/backend/internal/service"
	"net/http"

	"github.com/labstack/echo/v4"
)

type BrowserHandler struct {
	browserService service.BrowserService
}

func NewBrowserHandler(browserService service.BrowserService) *BrowserHandler {
	return &BrowserHandler{
		browserService: browserService,
	}
}

// ============================================
// ✅ LIST FILES (Handler)
// ============================================
// Endpoint: GET /api/v1/browser/files?remote=Gdrive1&path=/backups
// Response: BrowserResponse dengan list files
func (h *BrowserHandler) ListFiles(c echo.Context) error {
	// 1. Extract query parameters
	remoteName := c.QueryParam("remote")
	path := c.QueryParam("path")

	// 2. Validate parameters
	if remoteName == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "remote parameter is required",
		})
	}

	if path == "" {
		path = "/" // Default root path
	}

	// 3. Call service
	response, err := h.browserService.BrowseFiles(remoteName, path)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	// 4. Return response
	return c.JSON(http.StatusOK, response)
}

// ============================================
// ✅ GET FILE INFO (Handler)
// ============================================
// Endpoint: GET /api/v1/browser/info?remote=Gdrive1&file=/backups/file.zip
func (h *BrowserHandler) GetFileInfo(c echo.Context) error {
	remoteName := c.QueryParam("remote")
	filePath := c.QueryParam("file")

	if remoteName == "" || filePath == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "remote and file parameters are required",
		})
	}

	file, err := h.browserService.GetFileInfo(remoteName, filePath)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, file)
}

func (h *BrowserHandler) GetAvailableRemotes(c echo.Context) error {

	remotes, err := h.browserService.GetAvailableRemotes()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, remotes)
}
