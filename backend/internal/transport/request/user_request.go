package request

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email" binding:"required,email"`
	Avatar      string `json:"avatar"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

// ListUsersRequest 用户列表查询请求
type ListUsersRequest struct {
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"pageSize,default=15"`
	Username  string `form:"username"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
}
