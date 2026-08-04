package v1

var (
	// common errors
	ErrSuccess             = newError(0, "ok")
	ErrBadRequest          = newError(400, "Bad Request")
	ErrUnauthorized        = newError(401, "Unauthorized")
	ErrForbidden           = newError(403, "Forbidden")
	ErrNotFound            = newError(404, "Not Found")
	ErrInternalServerError = newError(500, "Internal Server Error")

	// more biz errors
	ErrEmailAlreadyUse     = newError(1001, "The email is already in use.")
	ErrInsufficientVoucher = newError(1002, "Insufficient contact voucher.")

	// enterprise errors
	ErrEnterpriseNotFound  = newError(2001, "Enterprise not found.")
	ErrEnterpriseDuplicate = newError(2002, "Enterprise already exists.")
	ErrEnterpriseOCRFailed = newError(2003, "OCR recognition failed.")
	ErrInvalidCreditCode   = newError(2004, "Invalid social credit code.")

	// job errors
	ErrShareRefreshLimitExceeded = newError(3001, "Share refresh already used today.")

	// upload errors
	ErrImageRiskyContent = newError(5001, "Image content may be risky.")

	// admin errors
	ErrAdminLoginFailed  = newError(4001, "Invalid admin username or password.")
	ErrAdminUnauthorized = newError(4002, "Admin token invalid or expired.")
)
