package money

import (
	"math"
	"regexp"
	"strings"

	"candypro/api/internal/pkg/valerr"
)

// MoneyEpsilon 浮点金额比较容差（与付款「剩余应付」等处一致）
const MoneyEpsilon = 1e-6

// MoneyCentTolerance 货币显示精度容差（2 位小数），避免 6695.81 付款 vs 6695.8125 订单总额无法标 paid
const MoneyCentTolerance = 0.01

// RoundMoney 将金额四舍五入到 2 位小数（订单/税/运费/付款统一精度）
func RoundMoney(amount float64) float64 {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return math.Round(amount*100) / 100
}

// MoneyCoversTotal 判断已确认付款累计是否达到或超过订单总额（避免 float 舍入导致无法标为 paid）
func MoneyCoversTotal(confirmedTotal, orderTotal float64) bool {
	// 排除 NaN/Inf，再用容差比较「已付是否覆盖应付」
	if math.IsNaN(confirmedTotal) || math.IsNaN(orderTotal) {
		return false
	}
	if math.IsInf(confirmedTotal, 0) || math.IsInf(orderTotal, 0) {
		return false
	}
	tolerance := MoneyCentTolerance
	if tolerance < MoneyEpsilon {
		tolerance = MoneyEpsilon
	}
	return confirmedTotal+tolerance >= orderTotal
}

// MoneyToCentsInt 将金额按货币精度（2 位小数）四舍五入后转为整数分
// （19.999 -> 2000，-19.99 -> -1999）。用于 lineage/审计哈希等需要金额
// 规范精确表示的序列化场景：以整数分参与哈希，避免 fmt.Sprintf("%f", ...)
// 保留 6 位小数的浮点尾噪导致「相同分值产生不同哈希 / 相邻分值碰撞」。
func MoneyToCentsInt(amount float64) int64 {
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		return 0
	}
	return int64(math.Round(amount * 100))
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
