package swagger

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"password123"`
}

type UserData struct {
	Username string `json:"username" example:"john_doe"`
}

type SuccessRegisterResponse struct {
	Code    int      `json:"code" example:"200"`
	Message string   `json:"message" example:"User successfully created"`
	Data    UserData `json:"data"`
}

type ServerErrorResponse struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"Something went wrong, please try again later"`
}

type ConflictResponse struct {
	Code    int    `json:"code" example:"409"`
	Message string `json:"message" example:"Username already exists"`
}

type ErrorResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid request data"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDI5NzU2MjgsImlkIjoiMTE3YzA4Y2EtZWEzNS00MWEyLWI4MDYtM2M5MmRjNTliMzhlIn0.cJ7xWY_V5dkIxrHfcPub--kUWZP4i2ky1nZDGkPL_BI"`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NDM1NzIxMzcsImlkIjoiODE4YmRmNGMtMGI5NC00ZGNiLTk2YmUtMTJhMzFmMDczYWMyIn0.5WCp11fVMXRKMzCzQvltEAC9sN_16u3AtUrMH7Z5JwI"`
}

type UnauthorizedResponse struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"Invalid credentials"`
}
