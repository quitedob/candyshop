package order

import "errors"

// ErrPaymentStateMismatch 更新付款状态时行数为 0（非预期状态或并发竞争）
var ErrPaymentStateMismatch = errors.New("payment state mismatch")

// ErrReturnStateMismatch 更新退货状态时行数为 0（非预期状态或并发竞争）
var ErrReturnStateMismatch = errors.New("return state mismatch")

// ErrFulfillmentStateMismatch 更新履约状态时行数为 0（非预期状态或并发竞争）
var ErrFulfillmentStateMismatch = errors.New("fulfillment state mismatch")

// ErrInvoiceStateMismatch 更新发票状态时行数为 0（非预期状态或并发竞争）
var ErrInvoiceStateMismatch = errors.New("invoice state mismatch")

// ErrOrderStateMismatch 更新订单状态时行数为 0（非预期状态或并发竞争）
var ErrOrderStateMismatch = errors.New("order state mismatch")
