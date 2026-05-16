package handler

import (
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/yourusername/wemake/internal/helper"
	mediautil "github.com/yourusername/wemake/internal/media"
)

type MediaHandler struct {
	publicBaseURL string
	cld           *cloudinary.Cloudinary
}

func NewMediaHandler(publicBaseURL string, cld *cloudinary.Cloudinary) *MediaHandler {
	return &MediaHandler{
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		cld:           cld,
	}
}

func (h *MediaHandler) UploadFile(c *fiber.Ctx) error {
	result, err := mediautil.SaveUploadedFile(c, mediautil.UploadOptions{
		FileNamePrefix:        uuid.NewString(),
		Folder:                "wemake",
		PublicBaseURL:         h.publicBaseURL,
		RequiredMessage:       "file is required in form-data",
		CloudUploadLogMessage: "cloudinary media upload failed",
		Cloudinary:            h.cld,
	})
	if err != nil {
		if uploadErr, ok := err.(*mediautil.UploadError); ok {
			return c.Status(uploadErr.Status).JSON(fiber.Map{"error": uploadErr.Message})
		}
		return helper.InternalServerError(c, "failed to save file")
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"url":       result.URL,
		"file_name": result.FileName,
		"size":      result.Size,
	})
}
