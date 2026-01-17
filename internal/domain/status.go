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

func (s StudentStatus) IsValid() bool {
	switch s {
	case StudentActive, StudentInactive, StudentSuspended:
		return true
	default:
		return false
	}
}

func (s SubscriptionStatus) IsValid() bool {
	switch s {
	case SubscriptionActive, SubscriptionEnded, SubscriptionCanceled, SubscriptionSuspended:
		return true
	default:
		return false
	}
}

func (s PaymentStatus) IsValid() bool {
	switch s {
	case PaymentConfirmed, PaymentReversed:
		return true
	default:
		return false
	}
}

func (s PaymentKind) IsValid() bool {
	switch s {
	case PaymentFull, PaymentPartial, PaymentAdvance, PaymentCredit:
		return true
	default:
		return false
	}
}

func (s PaymentMethod) IsValid() bool {
	switch s {
	case PaymentCash, PaymentPix, PaymentCard, PaymentTransfer, PaymentOther:
		return true
	default:
		return false
	}
}

func (s BillingPeriodStatus) IsValid() bool {
	switch s {
	case BillingOpen, BillingPaid, BillingPartial, BillingOverdue:
		return true
	default:
		return false
	}
}

func (s UserRole) IsValid() bool {
	switch s {
	case RoleAdmin, RoleOperator:
		return true
	default:
		return false
	}
}
