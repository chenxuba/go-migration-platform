package institutionmenu

import "strings"

const (
	legacyGroupCodePrefix = "INST_GROUP_"
	legacyRouteCodePrefix = "INST_ROUTE_"
	legacyAuthCodePrefix  = "INST_AUTH_"
)

func IsLegacyCode(code string) bool {
	trimmed := strings.TrimSpace(code)
	return strings.HasPrefix(trimmed, legacyGroupCodePrefix) ||
		strings.HasPrefix(trimmed, legacyRouteCodePrefix) ||
		strings.HasPrefix(trimmed, legacyAuthCodePrefix)
}

func LegacyToCurrentCode(code string) string {
	trimmed := strings.TrimSpace(code)
	switch {
	case strings.HasPrefix(trimmed, legacyGroupCodePrefix):
		return groupCodePrefix + upperSnakeToCompactCamel(strings.TrimPrefix(trimmed, legacyGroupCodePrefix))
	case strings.HasPrefix(trimmed, legacyRouteCodePrefix):
		return routeCodePrefix + upperSnakeToCompactCamel(strings.TrimPrefix(trimmed, legacyRouteCodePrefix))
	case strings.HasPrefix(trimmed, legacyAuthCodePrefix):
		return authCodePrefix + upperSnakeToCompactCamel(strings.TrimPrefix(trimmed, legacyAuthCodePrefix))
	default:
		return trimmed
	}
}
