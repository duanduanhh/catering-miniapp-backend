package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type OrderHandler struct {
	*Handler
	orderService service.OrderService
}

func NewOrderHandler(
    handler *Handler,
    orderService service.OrderService,
) *OrderHandler {
	return &OrderHandler{
		Handler:      handler,
		orderService: orderService,
	}
}

func (h *OrderHandler) GetOrder(ctx *gin.Context) {

}
