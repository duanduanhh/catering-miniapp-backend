package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type PositionCategoryHandler struct {
	*Handler
	positionCategoryService service.PositionCategoryService
}

func NewPositionCategoryHandler(
	handler *Handler,
	positionCategoryService service.PositionCategoryService,
) *PositionCategoryHandler {
	return &PositionCategoryHandler{
		Handler:                 handler,
		positionCategoryService: positionCategoryService,
	}
}

// GetAllCategories godoc
// @Summary 获取岗位分类列表
// @Tags 通用接口
// @Accept json
// @Produce json
// @Success 200 {object} v1.GetPositionCategoryResponseData
// @Router /positions/all [get]
func (h *PositionCategoryHandler) GetAllCategories(ctx *gin.Context) {
	categories, err := h.positionCategoryService.GetAllCategories(ctx)
	if err != nil {
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	categoriesList := make([]v1.PositionCategory, 0, len(categories))
	for _, category := range categories {
		tmp := v1.PositionCategory{
			ID:           category.ID,
			CategoryName: category.CategoryName,
		}
		for _, subcategory := range category.Subcategories {
			tmp.Subcategories = append(tmp.Subcategories, &v1.PositionSubcategory{
				ID:              subcategory.ID,
				SubcategoryName: subcategory.SubcategoryName,
				Description:     subcategory.Description,
			})
		}
		tmp.SubTotal = len(tmp.Subcategories)
		categoriesList = append(categoriesList, tmp)
	}

	v1.HandleSuccess(ctx,
		v1.GetPositionCategoryResponseData{
			List:  categoriesList,
			Total: len(categoriesList),
		},
	)
}
