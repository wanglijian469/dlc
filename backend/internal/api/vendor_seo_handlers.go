package api

import (
	"net/http"

	"dalu-nongji-parts/backend/internal/model"
	"dalu-nongji-parts/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func (h AdminHandler) SuggestVendorSEO(c *gin.Context) {
	var vendor model.Vendor
	if err := c.ShouldBindJSON(&vendor); err != nil {
		Fail(c, http.StatusBadRequest, 400, "厂商资料格式不正确")
		return
	}
	OK(c, service.SuggestVendorSEO(vendor))
}
