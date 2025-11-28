package handlers

import (
	"net/http"

	"ecommerce-order-product/usecase"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	uc *usecase.ProductUsecase
}

func NewProductHandler(uc *usecase.ProductUsecase) *ProductHandler {
	return &ProductHandler{uc}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	name := c.Query("name")
	category := c.Query("category")

	res, err := h.uc.GetProducts(name, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")

	res, err := h.uc.GetProduct(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, res)
}
