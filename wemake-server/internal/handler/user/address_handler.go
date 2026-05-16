package user

import (
	"database/sql"
	"errors"
	"github.com/yourusername/wemake/internal/dto"
	"github.com/yourusername/wemake/internal/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/yourusername/wemake/internal/domain"
	userservice "github.com/yourusername/wemake/internal/service/user"
	"github.com/yourusername/wemake/internal/domainutil"
)

type AddressHandler struct {
	service *userservice.AddressService
}

func NewAddressHandler(service *userservice.AddressService) *AddressHandler {
	return &AddressHandler{service: service}
}

func normalizeAddressType(raw string) (string, bool) {
	typ := domainutil.NormalizeStatus(raw)
	switch typ {
	case "B", "S", "C", "M":
		return typ, true
	default:
		return typ, false
	}
}

func (h *AddressHandler) ListAddresses(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_USER_ID", "invalid X-User-ID header"))
	}

	addresses, err := h.service.ListByUserID(userID)
	if err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("FETCH_ADDRESSES_FAILED", "failed to fetch addresses"))
	}
	return helper.WriteListResponse(c, addresses, len(addresses))
}

func (h *AddressHandler) CreateAddress(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_USER_ID", "invalid X-User-ID header"))
	}

	var req dto.CreateAddressRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}

	if strings.TrimSpace(req.AddressType) == "" || strings.TrimSpace(req.AddressDetail) == "" {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("MISSING_FIELDS", "address_type and address_detail are required"))
	}
	addressType, ok := normalizeAddressType(req.AddressType)
	if !ok {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_ADDRESS_TYPE", "address_type must be one of B, S, C, M"))
	}

	address := &domain.Address{
		UserID:        userID,
		AddressType:   addressType,
		AddressDetail: strings.TrimSpace(req.AddressDetail),
		SubDistrictID: req.SubDistrictID,
		DistrictID:    req.DistrictID,
		ProvinceID:    req.ProvinceID,
		ZipCode:       strings.TrimSpace(req.ZipCode),
		IsDefault:     req.IsDefault,
	}

	if err := h.service.Create(address); err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("CREATE_ADDRESS_FAILED", "failed to create address"))
	}
	c.Status(fiber.StatusCreated)
	return c.JSON(address)
}

func (h *AddressHandler) PatchAddress(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_USER_ID", "invalid X-User-ID header"))
	}

	addressID, err := helper.RequireInt64Param(c, "address_id")
		if err != nil {
			return err
		}

	var req dto.PatchAddressRequest
	if err := helper.RequireBody(c, &req); err != nil {
		return err
	}

	fields := map[string]interface{}{}
	if req.AddressType != nil {
		addressType, ok := normalizeAddressType(*req.AddressType)
		if !ok {
			return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_ADDRESS_TYPE", "address_type must be one of B, S, C, M"))
		}
		fields["address_type"] = addressType
	}
	if req.AddressDetail != nil {
		fields["address_detail"] = helper.DereferenceString(req.AddressDetail, "")
	}
	if req.SubDistrictID != nil {
		fields["sub_district_id"] = *req.SubDistrictID
	}
	if req.DistrictID != nil {
		fields["district_id"] = *req.DistrictID
	}
	if req.ProvinceID != nil {
		fields["province_id"] = *req.ProvinceID
	}
	if req.ZipCode != nil {
		fields["zip_code"] = helper.DereferenceString(req.ZipCode, "")
	}
	if req.IsDefault != nil {
		fields["is_default"] = *req.IsDefault
	}

	if err := h.service.Patch(userID, int64(addressID), fields); err != nil {
		return helper.WriteAPIError(c, helper.InternalServerError("PATCH_ADDRESS_FAILED", "failed to patch address"))
	}
	return helper.WriteSuccess(c, "address updated", nil)
}

func (h *AddressHandler) DeleteAddress(c *fiber.Ctx) error {
	userID, err := helper.UserIDFromHeader(c)
	if err != nil {
		return helper.WriteAPIError(c, helper.BadRequestAPIError("INVALID_USER_ID", "invalid X-User-ID header"))
	}
	addressID, err := helper.RequireInt64Param(c, "address_id")
		if err != nil {
			return err
		}
	if err := h.service.Delete(userID, int64(addressID)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return helper.WriteAPIError(c, helper.NotFoundAPIError("ADDRESS_NOT_FOUND", "address not found"))
		}
		return helper.WriteAPIError(c, helper.InternalServerError("DELETE_ADDRESS_FAILED", "failed to delete address"))
	}
	return c.SendStatus(fiber.StatusNoContent)
}
