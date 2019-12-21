package handlers

import (
	"github.com/rajendraventurit/radicaapi/lib/env"
	"github.com/rajendraventurit/radicaapi/lib/routetable"
)

// GetRoutes returns the combined routes
func GetRoutes(env *env.Env) routetable.RouteTable {
	rt := routetable.NewRouteTable()
	rt.Combine(
		UserRoutes(env),
	)
	return rt
}
