package swagger

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type UserRequest struct {
	Username string `json:"username" example:"john_doe"`
	Password string `json:"password" example:"password123"`
}

type TodoRequest struct {
	Title       string   `json:"title" example:"Buy groceries"`
	Description string   `json:"description" example:"Get milk, bread, and eggs"`
	Tags        []string `json:"tags" example:"shopping,urgent"`
	DueDate     string   `json:"due_date" example:"2025-04-01"`
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

type TodoResponse struct {
	ID          int      `json:"id" example:"12"`
	Title       string   `json:"title" example:"Buy groceries"`
	Description string   `json:"description" example:"Get milk, bread, and eggs"`
	Tags        []string `json:"tags" example:"shopping,urgent"`
	DueDate     string   `json:"due_date" example:"2025-04-01"`
}

type CreateTodoResponse struct {
	Code    int            `json:"code" example:"200"`
	Message string         `json:"message" example:"Successfully create"`
	Data    createResponse `json:"data"`
}

type createResponse struct {
	Id int `json:"id" example:"12"`
}

type SuccessTodoResponse struct {
	Code    int          `json:"code" example:"200"`
	Message string       `json:"message" example:"Successfully create"`
	Data    TodoResponse `json:"data"`
}

type SuccessEmptyResponse struct {
	Code    int    `json:"code" example:"200"`
	Message string `json:"message" example:"Successfully delete"`
}

type ListTodoResponse struct {
	Code    int            `json:"code" example:"200"`
	Message string         `json:"message" example:"Ok"`
	Offset  int            `json:"offset" example:"0"`
	Limit   int            `json:"limit" example:"20"`
	Count   int            `json:"count" example:"1"`
	Total   int            `json:"total" example:"1"`
	Results []TodoResponse `json:"data"`
}

type NotFoundResponse struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"Todo not found"`
}

type InvalidIDResponse struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"Invalid ID"`
}
