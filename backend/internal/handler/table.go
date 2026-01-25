package handler

import (
	"net/http"
	"strconv"
	"strings"

	"piemdm/internal/model"
	"piemdm/internal/service"
	"piemdm/pkg/helper/resp"

	"github.com/gin-gonic/gin"
)

type TableHandler interface {
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
}

type tableHandler struct {
	*Handler
	tableService           service.TableService
	tablePermissionService service.TablePermissionService
}

func NewTableHandler(handler *Handler, tableService service.TableService, tablePermissionService service.TablePermissionService) TableHandler {
	return &tableHandler{
		Handler:                handler,
		tableService:           tableService,
		tablePermissionService: tablePermissionService,
	}
}

// get /api/v1/tables
// get /api/v1/tables/:id
// post /api/v1/tables
// put /api/v1/tables/:id
// delete /api/v1/tables/:id
// post /api/v1/tables/batch
// put /api/v1/tables/batch
// delete /api/v1/tables/batch
// List 获取表列表
// @Summary 获取表列表
// @Tags 表管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(15)
// @Param code query string false "表代码"
// @Param startDate query string false "开始日期"
// @Param endDate query string false "结束日期"
// @Param relation query string false "关联关系"
// @Success 200 {array} model.Table
// @Router /admin/tables [get]
func (h *tableHandler) List(c *gin.Context) {
	// page=1&pageSize=15&showType=All&startDate=2023-01-01&endDate=2023-07-30
	page, pageSize := GetPage(c)
	where := make(map[string]any)
	var total int64

	code := c.Query("code")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	relation := c.Query("relation")

	if code != "" {
		where["code"] = code
	}
	if startDate != "" {
		where["created_at >="] = startDate
	}
	if endDate != "" {
		where["created_at <="] = endDate
	}
	if relation != "" {
		where["relation"] = relation
	}
	// h.logger.Info("ListTables where: ", "where", where)

	// Apply permission filtering
	userIDStr := c.GetString("user_id")
	if userIDStr != "" {
		// Convert user_id string to uint
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			h.logger.Error("Failed to parse user_id", "user_id", userIDStr, "error", err)
			resp.HandleError(c, http.StatusInternalServerError, "Invalid user ID", nil)
			return
		}

		allowedTableCodes, err := h.tablePermissionService.GetAllowedTableCodes(c, uint(userID))
		if err != nil {
			resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		// If allowedTableCodes is nil, user is superuser/admin (all tables allowed)
		// If it's an empty slice, user has no permissions
		// If it's a non-empty slice, filter by these codes
		if allowedTableCodes != nil {
			if len(allowedTableCodes) == 0 {
				// User has no table permissions
				resp.HandleSuccess(c, []any{})
				return
			}
			where["code"] = allowedTableCodes
		}
	}

	if pageSize == -1 {
		tables, err := h.tableService.Find("", where)
		if err != nil {
			resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
			return
		}
		resp.HandleSuccess(c, tables)
		return
	}

	for k, v := range c.Request.URL.Query() {
		// v[0] not enpty then use condition
		if v[0] != "" {
			// replace fullwidth character
			str := strings.ReplaceAll(v[0], "，", ",")
			where[k] = strings.TrimSpace(str)
		}
	}

	// delete control parameter
	delete(where, "page")
	delete(where, "pageSize")
	delete(where, "startDate")
	delete(where, "endDate")
	// h.logger.Info("where: ", "where", where)

	tables, err := h.tableService.List(page, pageSize, &total, where)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	links := resp.GeneratePaginationLinks(c.Request, page, pageSize, int(total))
	c.Header("Link", links.String())

	resp.HandleSuccess(c, tables)
}

// Get 获取表详情
// @Summary 获取表详情
// @Tags 表管理
// @Accept json
// @Produce json
// @Param id path int true "表ID"
// @Success 200 {object} model.Table
// @Router /admin/tables/{id} [get]
func (h *tableHandler) Get(c *gin.Context) {
	var req struct {
		Id uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	approval, err := h.tableService.Get(req.Id)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, approval)
}

// Create 创建表
// @Summary 创建表
// @Tags 表管理
// @Accept json
// @Produce json
// @Param data body model.Table true "表信息"
// @Success 200 {object} model.Table
// @Router /admin/tables [post]
func (h *tableHandler) Create(c *gin.Context) {
	var table model.Table
	// resultByte, _ := io.ReadAll(c.Request.Body)

	if err := c.ShouldBindJSON(&table); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	err := h.tableService.Create(c, &table)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	resp.HandleSuccess(c, table)
}

// Update 更新表
// @Summary 更新表
// @Tags 表管理
// @Accept json
// @Produce json
// @Param id path int true "表ID"
// @Param data body model.Table true "表信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/tables/{id} [put]
func (h *tableHandler) Update(c *gin.Context) {
	var table model.Table
	// resultByte, _ := io.ReadAll(c.Request.Body)

	if err := c.ShouldBindJSON(&table); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	// Prevent code modification after creation - clear code so GORM won't update it
	table.Code = ""

	err := h.tableService.Update(c, &table)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, nil)
}

func (h *tableHandler) Delete(c *gin.Context) {
	resp.HandleSuccess(c, nil)
}

func (h *tableHandler) BatchCreate(c *gin.Context) {
	resp.HandleSuccess(c, nil)
}

// BatchUpdate 批量更新表状态
// @Summary 批量更新表状态
// @Tags 表管理
// @Accept json
// @Produce json
// @Param data body object true "批量更新请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/tables/batch [put]
func (h *tableHandler) BatchUpdate(c *gin.Context) {
	var req struct {
		IDs    []uint `form:"ids"`
		Status string `form:"status"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	var table model.Table
	table.Status = req.Status

	err := h.tableService.BatchUpdate(c, req.IDs, &table)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
	// resp.HandleSuccess(c, nil)
}

// BatchDelete 批量删除表
// @Summary 批量删除表
// @Tags 表管理
// @Accept json
// @Produce json
// @Param data body object true "批量删除请求"
// @Success 200 {object} map[string]interface{}
// @Router /admin/tables/batch_delete [post]
func (h *tableHandler) BatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint `form:"ids"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.HandleError(c, http.StatusBadRequest, err.Error(), nil)
		return
	}
	err := h.tableService.BatchDelete(c, req.IDs)
	if err != nil {
		resp.HandleError(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	resp.HandleSuccess(c, gin.H{
		"message": "success",
	})
}
