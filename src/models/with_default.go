package models

func Int64WithDefault(val, defaultValue int64) int64 {
	if val == 0 {
		return defaultValue
	}
	return val
}
