package money

import (
	"math"
	"regexp"
	"strings"

	"candypro/api/internal/pkg/valerr"
)

// MoneyEpsilon 浮点金额比较容差（与付款「剩余应付」等处一致）
const MoneyEpsilon = 1e-6

// MoneyCoversTotal 判断已确认付款累计是否达到或超过订单总额（避免 float 舍入导致无法标为 paid）
func MoneyCoversTotal(confirmedTotal, orderTotal float64) bool {
	// 排除 NaN/Inf，再用容差比较「已付是否覆盖应付」
	if math.IsNaN(confirmedTotal) || math.IsNaN(orderTotal) {
		return false
	}
	if math.IsInf(confirmedTotal, 0) || math.IsInf(orderTotal, 0) {
		return false
	}
	return confirmedTotal+MoneyEpsilon >= orderTotal
}

var isoCurrencyRegexp = regexp.MustCompile(`^[A-Z]{3}$`)

// NormalizeISOCurrency 将币种规范为大写三位字母；空字符串表示未提供（由调用方默认 USD 等）
func NormalizeISOCurrency(raw string) (string, error) {
	s := strings.ToUpper(strings.TrimSpace(raw)) // 去空白并大写
	if s == "" {
		return "", nil
	}
	if !isoCurrencyRegexp.MatchString(s) {
		return "", valerr.ErrInvalidCurrencyISO
	}
	return s, nil
}
