package service

import (
	"github.com/yourusername/wemake/internal/domain"
	"github.com/yourusername/wemake/internal/logger"
	"github.com/yourusername/wemake/internal/repository"
	factoryrepo "github.com/yourusername/wemake/internal/repository/factory"
)

type AddressService struct {
	repo        *repository.AddressRepository
	factoryRepo *factoryrepo.FactoryRepository
}

func NewAddressService(repo *repository.AddressRepository, factoryRepo *factoryrepo.FactoryRepository) *AddressService {
	return &AddressService{repo: repo, factoryRepo: factoryRepo}
}

func (s *AddressService) ListByUserID(userID int64) ([]domain.Address, error) {
	return s.repo.ListByUserID(userID)
}

func (s *AddressService) Create(address *domain.Address) error {
	if err := s.repo.Create(address); err != nil {
		return err
	}
	// When a main (M-type) or default address is saved and has a province,
	// mirror province_id into factory_profiles so the factory location stays
	// in sync without requiring a separate profile update.
	if address.IsDefault && address.ProvinceID != 0 {
		if err := s.factoryRepo.PatchProfile(address.UserID, map[string]interface{}{
			"province_id": address.ProvinceID,
		}); err != nil {
			// Non-fatal: log the failure but don't roll back the address insert.
			logger.Warn("address province sync failed after create", "user_id", address.UserID, "province_id", address.ProvinceID, "err", err)
		}
	}
	return nil
}

func (s *AddressService) Patch(userID, addressID int64, fields map[string]interface{}) error {
	if err := s.repo.Patch(userID, addressID, fields); err != nil {
		return err
	}
	// If province_id was updated and address is being set as default,
	// keep factory_profiles.province_id in sync.
	provinceID, hasProvince := fields["province_id"]
	isDefault, hasDefault := fields["is_default"]
	if hasProvince && (!hasDefault || isDefault == true) {
		if pid, ok := provinceID.(int64); ok && pid != 0 {
			if err := s.factoryRepo.PatchProfile(userID, map[string]interface{}{
				"province_id": pid,
			}); err != nil {
				logger.Warn("address province sync failed after patch", "user_id", userID, "province_id", pid, "err", err)
			}
		}
	}
	return nil
}

func (s *AddressService) Delete(userID, addressID int64) error {
	return s.repo.Delete(userID, addressID)
}
