package vendors

// CommissionDefaultBps is the platform-wide default commission (10.00%).
// It is written into the vendor row at registration so the rate is frozen
// per-vendor from day one; admins override per vendor afterward.
//
// TODO(v1): surface as an admin setting instead of a constant when the
// settings schema gains a marketplace section.
const CommissionDefaultBps = 1000

// StaticCommissionDefaults implements CommissionDefaults with the constant.
type StaticCommissionDefaults struct{}

func NewStaticCommissionDefaults() *StaticCommissionDefaults {
	return &StaticCommissionDefaults{}
}

func (s *StaticCommissionDefaults) VendorCommissionBps() int {
	return CommissionDefaultBps
}

var _ CommissionDefaults = (*StaticCommissionDefaults)(nil)
