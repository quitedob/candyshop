package kyb

import (
	"math"
	"strings"
)

// UsdCapValue converts a monetary amount to its USD-equivalent using the
// configured exchange rates (base USD). Amounts already in USD pass through;
// empty currency is treated as USD. A currency without a configured rate fails
// closed (+Inf) so it can never satisfy a USD cap (M6: non-USD bypass). Callers
// must only compare the result as "less than a cap" — +Inf means "never allowed".
func UsdCapValue(rates map[string]float64, currency string, amount float64) float64 {
	cur := strings.ToUpper(strings.TrimSpace(currency))
	if cur == "" || cur == "USD" {
		return amount
	}
	rate, ok := rates[cur]
	if !ok || rate <= 0 {
		return math.Inf(1)
	}
	return amount / rate
}
