package payment

import "context"

// GatewayRequest carries the data needed for a payment operation.
type GatewayRequest struct {
	Amount       float64
	Currency     string
	Description  string
	OrderID      string
	PaymentID    string
	IdempotencyKey string // optional: prevents duplicate operations
	Metadata     map[string]string
}

// GatewayResponse is returned by payment gateway operations.
type GatewayResponse struct {
	TransactionID string
	Status        string
	RawResponse   string // JSON dump of gateway response
	ClientSecret  string // Stripe PaymentIntent client_secret
	ApprovalURL   string // PayPal checkout approval URL
}

// PaymentGateway abstracts a payment processor (Stripe, PayPal, etc.).
type PaymentGateway interface {
	// Authorize reserves funds on the customer's payment method without capturing.
	Authorize(ctx context.Context, req GatewayRequest) (*GatewayResponse, error)
	// Capture captures previously authorized funds.
	Capture(ctx context.Context, transactionID string, amount float64) (*GatewayResponse, error)
	// Refund returns funds to the customer.
	Refund(ctx context.Context, transactionID string, amount float64) (*GatewayResponse, error)
	// Void cancels an authorization before capture.
	Void(ctx context.Context, transactionID string) (*GatewayResponse, error)
	// Name returns a human-readable gateway name for logging.
	Name() string
}
