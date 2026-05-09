package order

import "errors"

// ErrPaymentStateMismatch 更新付款状态时行数为 0（非预期状态或并发竞争）
var ErrPaymentStateMismatch = errors.New("payment state mismatch")
