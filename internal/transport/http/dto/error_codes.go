package dto

const (
	ErrTitlePathNotFound            = "path not found"
	ErrTitleResourceDoesntExist     = "resource does not exist"
	ErrTitleResourceAlreadyExists   = "resource already exists"
	ErrTitleInvalidData             = "invalid data"
	ErrTitleValidation              = "validation error"
	ErrTitleInvalidAuthData         = "invalid authentication data"
	ErrTitleInternal                = "internal server error"
	ErrTitleUnauthorized            = "unauthorized"
	ErrTitleTokenExpired            = "token expired"
	ErrTitleInvalidHash             = "invalid hash"
	ErrTitleCodeAlreadySent         = "code already sent"
	ErrTitleInvalidVerificationCode = "invalid verification code"
	ErrTitleUserIsBlocked           = "user is blocked"
	ErrTitleProfileNotCompleted     = "profile not completed"
	ErrTitleInvalidFileType         = "invalid file type"
	ErrTitleUserGeolocationNotSet   = "user geolocation is not set"
	ErrTitleNotMatch                = "have not match"
	ErrTitleMatchAlreadyExists      = "match already exists"
	ErrTitleBoostAlreadyExists      = "boost already exists"

	// ErrTitleNotSupportedPhone       = "not supported phone number" // зарезервировано на будущее
	// ErrTitleTooManyRequestsIP       = "ip is blocked"              // зарезервировано на будущее
	// ErrCodeNotSupportedPhone       = 1013 // зарезервировано на будущее
	// ErrCodeTooManyRequestsIP       = 1014 // зарезервировано на будущее

	ErrCodePathNotFound            = 1000
	ErrCodeResourceDoesntExist     = 1001
	ErrCodeResourceAlreadyExists   = 1002
	ErrCodeInvalidData             = 1003
	ErrCodeValidation              = 1004
	ErrCodeInvalidAuthData         = 1005
	ErrCodeInternal                = 1006
	ErrCodeUnauthorized            = 1007
	ErrCodeTokenExpired            = 1008
	ErrCodeInvalidHash             = 1009
	ErrCodeCodeAlreadySent         = 1010
	ErrCodeInvalidVerificationCode = 1011
	ErrCodeUserIsBlocked           = 1012
	ErrCodeProfileContCompleted    = 1015
	ErrCodeInvalidFileType         = 1016
	ErrCodeUserGeolocationNotSet   = 1017
	ErrCodeNotMatch                = 1018
	ErrCodeMatchAlreadyExists      = 1019
	ErrCodeBoostAlreadyExists      = 1020
)
