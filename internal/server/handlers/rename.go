package handlers

import (
	"net/http"
	"strconv"

	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/resp"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/store"

	"github.com/gin-gonic/gin"
)

func init() {
	router.NewGroupRouter("/api/v1/rename").
		Use(middleware.Auth()).
		AddRoute(
			router.NewRoute("/list", http.MethodGet).
				Handle(renameTemplateList),
		).
		AddRoute(
			router.NewRoute("/create", http.MethodPost).
				Handle(renameTemplateCreate),
		).
		AddRoute(
			router.NewRoute("/update/:id", http.MethodPut).
				Handle(renameTemplateUpdate),
		).
		AddRoute(
			router.NewRoute("/del/:id", http.MethodDelete).
				Handle(renameTemplateDelete),
		).
		AddRoute(
			router.NewRoute("/preview", http.MethodPost).
				Handle(renamePreview),
		)
}

func renameTemplateList(c *gin.Context) {
	resp.Success(c, store.RenameTemplateList())
}

// renameTemplateCreate 由服务端生成可信预览后保存模板。
func renameTemplateCreate(c *gin.Context) {
	var req struct {
		Expression string `json:"expression" binding:"required"` // 交给 Go 模板解析的重命名表达式。
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	preview, err := renderRenamePreview(req.Expression)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	template := model.RenameTemplate{Preview: preview, Expression: req.Expression}
	if err := store.RenameTemplateCreate(&template); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, template)
}

// renameTemplateUpdate 重新校验表达式，并用服务端生成的预览更新模板。
func renameTemplateUpdate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		resp.Error(c, http.StatusBadRequest, resp.ErrBadRequest)
		return
	}
	if _, ok := store.RenameTemplateGet(id); !ok {
		resp.Error(c, http.StatusNotFound, resp.ErrResourceNotFound)
		return
	}
	var req struct {
		Expression string `json:"expression" binding:"required"` // 交给 Go 模板解析的重命名表达式。
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}
	preview, err := renderRenamePreview(req.Expression)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	template := model.RenameTemplate{ID: id, Preview: preview, Expression: req.Expression}
	if err := store.RenameTemplateUpdate(&template); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, template)
}

func renameTemplateDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		resp.Error(c, http.StatusBadRequest, resp.ErrBadRequest)
		return
	}
	if err := store.RenameTemplateDelete(id); err != nil {
		resp.Error(c, http.StatusInternalServerError, resp.ErrDatabase)
		return
	}
	resp.Success(c, "rename template deleted successfully")
}

// renamePreview 使用固定节点数据校验表达式并返回与正式重命名一致的渲染结果。
func renamePreview(c *gin.Context) {
	var req struct {
		Expression string `json:"expression" binding:"required"` // 交给 Go 模板解析的重命名表达式。
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
		return
	}

	result, err := renderRenamePreview(req.Expression)
	if err != nil {
		resp.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	resp.Success(c, gin.H{"result": result})
}

// renderRenamePreview 使用固定节点生成可信示例，保证预览和保存结果一致。
func renderRenamePreview(expression string) (string, error) {
	return node.Rename(node.Node{
		Info: model.NodeInfo{Delay: 123, DownloadSpeed: 10240, CountryCode: "CN", TrafficMultiplier: 0.5},
	}, 1, expression)
}
