package handler

import (
	"net/http"

	"piemdm/internal/model"
	"piemdm/internal/transport/request"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"
	"piemdm/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type UserHandler interface {
	// Base CRUD
	List(c *gin.Context)
	Get(c *gin.Context)
	Create(c *gin.Context)
	Update(c *gin.Context)
	Delete(c *gin.Context)

	// Batch operations
	BatchCreate(c *gin.Context)
	BatchUpdate(c *gin.Context)
	BatchDelete(c *gin.Context)

	// 特殊操作
	Login(c *gin.Context)
	ValidateToken(c *gin.Context)

	// 角色管理
	GetUserRoles(c *gin.Context)
	UpdateUserRoles(c *gin.Context)
}

func NewUserHandler(handler *Handler, userService service.UserService, roleService service.RoleService, jwt *jwt.JWT) UserHandler {
	return &userHandler{
		Handler:     handler,
		userService: userService,
		roleService: roleService,
		jwt:         jwt,
	}
}

type userHandler struct {
	*Handler
	userService service.UserService
	roleService service.RoleService
	jwt         *jwt.JWT
}

// Login 用户登录
// @Summary 用户登录
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body request.LoginRequest true "登录信息"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *userHandler) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, errors.Wrap(err, "invalid request").Error(), nil)
		return
	}

	token, user, err := h.userService.Login(c, &req)
	if err != nil {
		resp.HandleError(c, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	// 只返回userid、username、avatar
	userInfo := gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"admin":    user.Admin,
		"avatar":   user.Avatar,
	}

	resp.HandleSuccess(c, gin.H{
		"token":    token,
		"userInfo": userInfo,
		// "roles": roles,
	})
}

func (h *userHandler) UpdateStatus(c *gin.Context) {
	var req struct {
		Ids    []uint `form:"ids"`
		Status string `form:"status"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var user model.User
	user.Status = req.Status

	err := h.userService.BatchUpdate(c, req.Ids, &user)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

// Create 创建用户
// @Summary 创建用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body object true "用户信息"
// @Success 200 {object} model.User
// @Router /admin/users [post]
func (h *userHandler) Create(c *gin.Context) {
	var req struct {
		EmployeeID  string `binding:"max=64"`
		Username    string `binding:"max=64"`
		Password    string `binding:"max=128"`
		FirstName   string `binding:"max=64"`
		LastName    string `binding:"max=64"`
		DisplayName string `binding:"max=64"`
		Email       string `binding:"max=255"`
		Phone       string `binding:"max=64"`
		Language    string `binding:"max=8"`
		Sex         string `binding:"max=8,oneof=Male Female Other"`
		Avatar      string `binding:"max=255"`
		Description string `binding:"max=255"`
		Admin       string `binding:"max=3,oneof=Yes No"`
		Status      string `binding:"oneof=Normal Frozen"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user := model.User{
		EmployeeID:  req.EmployeeID,
		Username:    req.Username,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Phone:       req.Phone,
		Language:    req.Language,
		Sex:         req.Sex,
		Avatar:      req.Avatar,
		Description: req.Description,
		Admin:       req.Admin,
		Status:      req.Status,
	}

	err := h.userService.Create(c, &user)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, user)
}

// List 获取用户列表
// @Summary 获取用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(15)
// @Success 200 {array} model.User
// @Router /admin/users [get]
func (h *userHandler) List(c *gin.Context) {
	var req request.ListUsersRequest

	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	users, total, err := h.userService.List(c, &req)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, req.Page, req.PageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, users)
}

// Get 获取用户详情
// @Summary 获取用户详情
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} model.User
// @Router /admin/users/{id} [get]
func (h *userHandler) Get(c *gin.Context) {
	var req struct {
		ID uint `uri:"id"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := h.userService.Get(req.ID)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, user)
}

// Update 更新用户
// @Summary 更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param data body object true "用户信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/users/{id} [put]
func (h *userHandler) Update(c *gin.Context) {
	var req struct {
		ID          uint   `binding:"required"`
		EmployeeID  string `binding:"max=64"`
		Username    string `binding:"max=64"`
		Password    string `binding:"max=128"`
		FirstName   string `binding:"max=64"`
		LastName    string `binding:"max=64"`
		DisplayName string `binding:"max=64"`
		Email       string `binding:"max=255"`
		Phone       string `binding:"max=64"`
		Language    string `binding:"max=8"`
		Sex         string `binding:"max=8,oneof=Male Female Other"`
		Avatar      string `binding:"max=255"`
		Description string `binding:"max=255"`
		Admin       string `binding:"max=3,oneof=Yes No"`
		Status      string `binding:"oneof=Normal Frozen"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user := model.User{
		ID:          req.ID,
		EmployeeID:  req.EmployeeID,
		Username:    req.Username,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Phone:       req.Phone,
		Language:    req.Language,
		Sex:         req.Sex,
		Avatar:      req.Avatar,
		Description: req.Description,
		Admin:       req.Admin,
		Status:      req.Status,
	}

	err3 := h.userService.Update(c, &user)
	if err3 != nil {
		resp.HandleError(c, http.StatusInternalServerError, err3.Error(), nil)
		return
	}
	resp.HandleSuccess(c, nil)
}

// Delete 删除用户
// @Summary 删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} map[string]interface{}
// @Router /admin/users/{id} [delete]
func (h *userHandler) Delete(c *gin.Context) {
	var req struct {
		ID uint `form:"id"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	err := h.userService.Delete(c, req.ID)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

func (h *userHandler) BatchCreate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Batch Create user"})
}

// BatchUpdate 批量更新用户状态
// @Summary 批量更新用户状态
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body object true "批量更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/users/batch [put]
func (h *userHandler) BatchUpdate(c *gin.Context) {
	var params struct {
		Ids    []uint `form:"ids"`
		Status string `form:"status"`
	}
	if err := c.ShouldBind(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var user model.User
	user.Status = params.Status

	err := h.userService.BatchUpdate(c, params.Ids, &user)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

// BatchDelete 批量删除用户
// @Summary 批量删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param data body object true "批量删除请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/users/batch [delete]
func (h *userHandler) BatchDelete(c *gin.Context) {
	var params struct {
		Ids []uint `form:"ids"`
	}
	if err := c.ShouldBind(&params); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	err := h.userService.BatchDelete(c, params.Ids)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}

// ValidateToken 验证Token
// @Summary 验证Token
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Success 200 {object} map[string]interface{}
// @Router /auth/validate [post]
func (h *userHandler) ValidateToken(c *gin.Context) {
	// 从header获取token
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		resp.HandleError(c, http.StatusUnauthorized, "缺少token", nil)
		return
	}

	// 校验token
	claims, err := h.jwt.ParseToken(tokenString)
	if err != nil {
		resp.HandleError(c, http.StatusUnauthorized, "token无效或已过期", nil)
		return
	}
	// 可选：返回用户基本信息
	resp.HandleSuccess(c, gin.H{
		"id": claims.ID,
	})
}

// GetUserRoles 获取用户的角色
// GetUserRoles 获取用户角色
// @Summary 获取用户角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {array} model.Role
// @Router /admin/users/{id}/roles [get]
func (h *userHandler) GetUserRoles(c *gin.Context) {
	var req struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	roles, err := h.userService.GetUserRoles(req.ID)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, roles)
}

// UpdateUserRoles 更新用户的角色
// UpdateUserRoles 更新用户角色
// @Summary 更新用户角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param data body object true "角色ID列表"
// @Success 200 {object} map[string]interface{}
// @Router /admin/users/{id}/roles [put]
func (h *userHandler) UpdateUserRoles(c *gin.Context) {
	var req struct {
		ID      uint   `uri:"id" binding:"required"`
		RoleIDs []uint `json:"role_ids" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	if err := h.userService.UpdateUserRoles(req.ID, req.RoleIDs); err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, gin.H{"message": "success"})
}
