package api

import (
	"fmt"
	"strings"

	"dalu-nongji-parts/backend/internal/model"
)

const maxVendorServiceAdvantagesRunes = 80

func validateVendorProfileContent(vendor model.Vendor) error {
	if len([]rune(strings.TrimSpace(vendor.ServiceAdvantages))) > maxVendorServiceAdvantagesRunes {
		return fmt.Errorf("服务优势不能超过 %d 个字符", maxVendorServiceAdvantagesRunes)
	}
	return nil
}
