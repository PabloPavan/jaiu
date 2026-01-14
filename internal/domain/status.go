package domain

type StudentStatus string

type SubscriptionStatus string

type PaymentStatus string

type PaymentMethod string

type PaymentKind string

type BillingPeriodStatus string

type UserRole string

const (
	StudentActive    StudentStatus = "active"
	StudentInactive  StudentStatus = "inactive"
	StudentSuspended StudentStatus = "suspended"
)

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionEnded     SubscriptionStatus = "ended"
	SubscriptionCanceled  SubscriptionStatus = "canceled"
	SubscriptionSuspended SubscriptionStatus = "suspended"
)

const (
	PaymentConfirmed PaymentStatus = "confirmed"
	PaymentReversed  PaymentStatus = "reversed"
)

const (
	PaymentFull    PaymentKind = "full"
	PaymentPartial PaymentKind = "partial"
	PaymentAdvance PaymentKind = "advance"
	PaymentCredit  PaymentKind = "credit"
)

const (
	PaymentCash     PaymentMethod = "cash"
	PaymentPix      PaymentMethod = "pix"
	PaymentCard     PaymentMethod = "card"
	PaymentTransfer PaymentMethod = "transfer"
	PaymentOther    PaymentMethod = "other"
)

const (
	BillingOpen    BillingPeriodStatus = "open"
	BillingPaid    BillingPeriodStatus = "paid"
	BillingPartial BillingPeriodStatus = "partial"
	BillingOverdue BillingPeriodStatus = "overdue"
)

const (
	RoleAdmin    UserRole = "admin"
	RoleOperator UserRole = "operator"
)
