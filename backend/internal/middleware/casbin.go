package middleware

import (
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
)

// CasbinMiddleware enforces permissions for the given resource and action.
// It assumes that the user ID is available in the gin context under the key "userId".
func CasbinMiddleware(enforcer *casbin.Enforcer, resource string, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current user ID
		userID, exists := c.Get("user_id")
		// Also support "user_id" key if used elsewhere, but strict usage is better.
		// auth_middleware.go sets "user_id" (from JWT).

		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No user ID found"})
			return
		}

		// Convert userID to string
		sub := fmt.Sprintf("%v", userID)

		// Skip check for superuser role if enforced via policy, but usually Enforcer handles it via matcher.
		// My matcher: m = g(r.sub, p.sub) && ... || r.sub == "superuser" ? No, I used "root" in example.
		// I should rely on policy logic.
		// If "g, 1, superuser" exists, and "p, superuser, *, *" exists (not standard regex support in basic model unless using regex matching).
		// Standard RBAC usually requires explicit permissions.

		// Enforce
		ok, err := enforcer.Enforce(sub, resource, action)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Casbin enforcement failed"})
			return
		}

		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message":  fmt.Sprintf("权限不足：您没有权限执行此操作，需要 %s 资源的 %s 权限", resource, action),
				"resource": resource,
				"action":   action,
			})
			return
		}

		c.Next()
	}
}
