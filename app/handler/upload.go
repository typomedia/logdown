package handler

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

// uploadForm renders the dropzone page. Ports UploadController::index.
func (h *Handler) uploadForm(c *fiber.Ctx) error {
	return h.render(c, h.page("app_upload_index"))
}

// upload accepts the uploaded logfiles and rebuilds the database from scratch.
// Ports UploadController::upload.
func (h *Handler) upload(c *fiber.Ctx) error {
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "logdown-upload-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	var paths []string
	for _, headers := range form.File {
		for _, fh := range headers {
			dst := filepath.Join(tmpDir, filepath.Base(fh.Filename))
			if err := c.SaveFile(fh, dst); err != nil {
				return err
			}
			paths = append(paths, dst)
		}
	}

	imported, err := h.repo.Rebuild(paths)
	if err != nil {
		return err
	}
	if imported == nil {
		imported = []string{}
	}
	return c.JSON(fiber.Map{"files": imported})
}
