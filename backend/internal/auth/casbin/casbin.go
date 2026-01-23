package casbin

import (
	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"gorm.io/gorm"
)

// InitEnforcer initializes the Casbin enforcer with the Gorm adapter.
func InitEnforcer(db *gorm.DB) (*casbin.Enforcer, error) {
	adapter := NewAdapter(db)

	// Define RBAC model
	text := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act || r.sub == "root"
`
	// Added r.sub == "root" just in case, but usually we use roles.
	// Using simple model first.

	m, err := model.NewModelFromString(text)
	if err != nil {
		return nil, err
	}

	e, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}

	// Load policies from DB
	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}

	return e, nil
}
