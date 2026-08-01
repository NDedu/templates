package validator

const (
	// General
	InvalidJSONErr        = "Invalid JSON"
	InvalidCredentialsErr = "Invalid credentials"
	InternalErr           = "Internal error"
	ValidationFailedErr   = "Validation failed"
	ServerBusyErr         = "Server is busy"
	RateLimitErr          = "Rate limit exceeded, try again later"
	InvalidIDErr          = "Invalid ID"

	// Auth / access
	InvalidOrExpiredTokenErr  = "Invalid or expired token"
	AuthRequiredErr           = "Authentication required"
	AdminRequiredErr          = "Admin access required"
	ForbiddenErr              = "Forbidden"
	AccountNotActiveResendErr = "Account not active, a new activation link has been sent to your email"

	// User fields
	EmailRequiredErr          = "Email required"
	EmailAlreadyRegisteredErr = "Email already registered"
	EmailsDoNotMatchErr       = "Emails do not match"
	PasswordRequiredErr       = "Password required"
	PasswordTooShortErr       = "Password too short, minimum 8 characters"
	PasswordTooLongErr        = "Password too long, maximum 72 characters"
	PasswordsDoNotMatchErr    = "Passwords do not match"
	FirstNameRequiredErr      = "First name is required"
	LastNameRequiredErr       = "Last name is required"
	InvalidRoleErr            = "Invalid role"

	// Users
	UserNotFoundErr     = "User not found"
	RetrieveUsersErr    = "Failed to retrieve users"
	CreateUserErr       = "Failed to create user"
	UpdateUserErr       = "Failed to update user"
	DeleteUserErr       = "Failed to delete user"
	CannotDeleteSelfErr = "You cannot delete your own account"

	// Profile image
	UploadImageErr      = "Failed to upload image"
	InvalidImageTypeErr = "Only JPEG, PNG, GIF, and WebP images are allowed"
	ImageTooLargeErr    = "Image must be less than 5MB"

	// Books
	SearchFailedErr     = "Search failed"
	BookNotFoundErr     = "Book not found"
	BookNameRequiredErr = "Book name is required"
	RetrieveBooksErr    = "Failed to retrieve books"
	CreateBookErr       = "Failed to create book"
	UpdateBookErr       = "Failed to update book"
	DeleteBookErr       = "Failed to delete book"
)
