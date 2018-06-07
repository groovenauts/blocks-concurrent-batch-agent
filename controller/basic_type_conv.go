package controller

func BoolPointerToBool(v *bool) bool {
	return BoolPointerToBoolWith(v, false)
}

func BoolPointerToBoolWith(v *bool, d bool) bool {
	if v == nil {
		return d
	}
	return *v
}

func IntPointerToInt(v *int) int {
	return IntPointerToIntWith(v, 0)
}

func IntPointerToIntWith(v *int, d int) int {
	if v == nil {
		return d
	}
	return *v
}

func StringPointerToString(v *string) string {
	return StringPointerToStringWith(v, "")
}

func StringPointerToStringWith(v *string, d string) string {
	if v == nil {
		return d
	}
	return *v
}
