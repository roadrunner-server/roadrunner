package utils

func AsUint64Ptr(val uint64) *uint64 {
	if val == 0 {
		val = 10
	}
	return &val
}

func AsStringPtr(val string) *string {
	if val == "" {
		return nil
	}
	return &val
}
