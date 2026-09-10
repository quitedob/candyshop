package trade

import "errors"

// ErrTradeStateMismatch 更新贸易主单状态时行数为 0（非预期状态或并发竞争）
var ErrTradeStateMismatch = errors.New("trade state mismatch")

// ErrTradeDocumentStateMismatch 更新贸易单证状态时行数为 0（非预期状态或并发竞争）
var ErrTradeDocumentStateMismatch = errors.New("trade document state mismatch")

// ErrShipmentStateMismatch 更新发货状态时行数为 0（非预期状态或并发竞争）
var ErrShipmentStateMismatch = errors.New("shipment state mismatch")

// ErrSettlementStateMismatch 更新结算状态时行数为 0（非预期状态或并发竞争）
var ErrSettlementStateMismatch = errors.New("settlement state mismatch")
