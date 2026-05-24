package dto

// Profile Request DTOs
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Email     string `json:"email"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"notblank"`
	NewPassword string `json:"new_password" validate:"notblank"`
}

type UpdateNotificationPreferencesRequest struct {
	EmailNotifications *bool `json:"email_notifications"`
	SMSNotifications   *bool `json:"sms_notifications"`
	PushNotifications  *bool `json:"push_notifications"`
	OrderUpdates       *bool `json:"order_updates"`
	PromotionalEmails  *bool `json:"promotional_emails"`
	NewsletterEmails   *bool `json:"newsletter_emails"`
}

// Address Request DTOs
type CreateAddressRequest struct {
	AddressType   string `json:"address_type"`
	AddressDetail string `json:"address_detail"`
	SubDistrictID int64  `json:"sub_district_id"`
	DistrictID    int64  `json:"district_id"`
	ProvinceID    int64  `json:"province_id"`
	ZipCode       string `json:"zip_code"`
	IsDefault     bool   `json:"is_default"`
}

type PatchAddressRequest struct {
	AddressType   *string `json:"address_type"`
	AddressDetail *string `json:"address_detail"`
	SubDistrictID *int64  `json:"sub_district_id"`
	DistrictID    *int64  `json:"district_id"`
	ProvinceID    *int64  `json:"province_id"`
	ZipCode       *string `json:"zip_code"`
	IsDefault     *bool   `json:"is_default"`
}
