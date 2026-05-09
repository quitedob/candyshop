package utils

import "errors"

// ErrInvalidCurrencyISO 币种必须为三位大写字母 ISO（如 USD）
var ErrInvalidCurrencyISO = errors.New("currency must be a 3-letter ISO code (e.g. USD)")
