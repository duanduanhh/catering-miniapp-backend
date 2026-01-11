package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type OrderItemHandler struct {
	*Handler
	orderItemService service.OrderItemService
}

func NewOrderItemHandler(
    handler *Handler,
    orderItemService service.OrderItemService,
) *OrderItemHandler {
	return &OrderItemHandler{
		Handler:      handler,
		orderItemService: orderItemService,
	}
}

func (h *OrderItemHandler) GetOrderItem(ctx *gin.Context) {

}
