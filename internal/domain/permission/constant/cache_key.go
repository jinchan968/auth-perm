package constant

import cacheConstant "auth-perm/internal/domain/cache/constant"

const (
	PermissionCacheKeyPrefix = "perm:permission:"
	ResourceCacheKeyPrefix   = "perm:resource:"
)

func PermissionCacheKey(userID string) string {
	return cacheConstant.BuildKey(PermissionCacheKeyPrefix, userID)
}

func ResourceCacheKey(accountID, resourceType string) string {
	return cacheConstant.BuildKey(ResourceCacheKeyPrefix, accountID, resourceType)
}
