package handler

import (
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/repository"
)

var errNotFound = sql.ErrNoRows

type serviceErrorCase struct {
	Err     error
	Status  int
	Message string
}

func writeServiceError(c *fiber.Ctx, err error, fallback string, cases ...serviceErrorCase) error {
	for _, item := range cases {
		if item.Err != nil && errors.Is(err, item.Err) {
			return jsonError(c, item.Status, item.Message)
		}
	}
	if errors.Is(err, domain.ErrForbidden) {
		return jsonError(c, fiber.StatusForbidden, "forbidden")
	}
	if repository.IsNotFoundError(err) {
		return jsonError(c, fiber.StatusNotFound, fallbackNotFoundMessage(fallback))
	}
	return jsonError(c, fiber.StatusInternalServerError, fallback)
}

func fallbackNotFoundMessage(fallback string) string {
	switch fallback {
	case "failed to fetch order", "failed to verify order", "failed to cancel order", "failed to confirm receipt", "failed to fetch review state", "failed to create review":
		return "order not found"
	case "failed to fetch quotation", "failed to authorize":
		return "quotation not found"
	case "failed to fetch factory", "failed to update factory":
		return "factory not found"
	default:
		return "not found"
	}
}

func badRequestCase(err error) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusBadRequest, Message: err.Error()}
}

func conflictCase(err error) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusConflict, Message: err.Error()}
}

func notFoundCase(err error, message string) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusNotFound, Message: message}
}

func unprocessableCase(err error) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusUnprocessableEntity, Message: err.Error()}
}

func forbiddenCase(err error) serviceErrorCase {
	return serviceErrorCase{Err: err, Status: fiber.StatusForbidden, Message: err.Error()}
}
