package routetable

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/bouk/httprouter"
)

// Route is an endpoint route
type Route struct {
	Name        string
	Description string
	Category    string
	Input       string
	Output      string
	Method      string
	Path        string
	Handler     http.Handler
	Insecure    bool
	Permissions []int64
}

// RouteTable is a collection of routes
type RouteTable struct {
	Routes []Route
	hashed map[string]*Route
}

// NewRouteTable returns a new route table
func NewRouteTable() RouteTable {
	return RouteTable{hashed: make(map[string]*Route)}
}

// Add will add a route to the route table
func (rt *RouteTable) Add(routes ...Route) {
	for _, r := range routes {
		rt.Routes = append(rt.Routes, r)
	}
	rt.hash()
}

func (rt *RouteTable) hash() {
	rt.hashed = make(map[string]*Route)
	for i, r := range rt.Routes {
		key := fmt.Sprintf("%s:%s", strings.ToUpper(r.Method), strings.ToLower(r.Path))
		rt.hashed[key] = &rt.Routes[i]
	}
}

// GetRoute will return  a matching route
func (rt RouteTable) GetRoute(method, path string) (*Route, error) {
	key := fmt.Sprintf("%s:%s", strings.ToUpper(method), strings.ToLower(path))
	if r, ok := rt.hashed[key]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("route not found")
}

// Methods returns the available methods for a path
func (rt RouteTable) Methods(path string) string {
	path = strings.ToLower(path)
	allow := []string{"OPTIONS"}
	for _, r := range rt.Routes {
		rpath := strings.ToLower(r.Path)
		if path == rpath {
			allow = append(allow, r.Method)
		}
	}
	return strings.Join(allow, ", ")
}

// Combine will add routes to this route table
func (rt *RouteTable) Combine(newrt ...RouteTable) {
	for _, r := range newrt {
		rt.Routes = append(rt.Routes, r.Routes...)
	}
	rt.hash()
}

// Router is an http router
type Router struct {
	Router     *httprouter.Router
	RouteTable RouteTable
	Middleware []func(http.Handler) http.Handler
}

// NewRouter returns a router
func NewRouter() *Router {
	r := httprouter.New()
	r.HandleOPTIONS = true
	r.RedirectTrailingSlash = true
	r.RedirectFixedPath = true
	r.NotFound = http.HandlerFunc(handle404)
	router := Router{Router: r}
	return &router
}

// AddMiddleware adds a middleware
func (r *Router) AddMiddleware(fn func(http.Handler) http.Handler) {
	r.Middleware = append(r.Middleware, fn)
}

// SetRouteTable sets the route table for the router
func (r *Router) SetRouteTable(rt RouteTable) {
	r.RouteTable = rt
}

func (r *Router) addMiddleware() http.Handler {
	fn := http.Handler(r.Router)
	for _, h := range r.Middleware {
		fn = h(fn)
	}
	return fn
}

// Handler returns a handler ready to serve routes
func (r Router) Handler() http.Handler {
	for _, route := range r.RouteTable.Routes {
		fn := http.Handler(route.Handler)
		r.Router.Handle(route.Method, route.Path, wrapHandler(fn))
	}
	return r.addMiddleware()
}

func wrapHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func handle404(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ERROR] %v %v 404 - Not Found", r.Method, r.URL)
	http.Error(w, "Not Found", http.StatusNotFound)
}

// GenMDDocumentation will generate documentation for each route
func (rt RouteTable) GenMDDocumentation() string {
	dm := make(map[string]string)
	for _, r := range rt.Routes {
		cat := r.Category
		if cat == "" {
			cat = "Unknown"
		}
		dlist := dm[cat]
		dlist += r.GenMDDocumentation()
		dm[cat] = dlist
	}

	cats := []string{}
	for k := range dm {
		cats = append(cats, k)
	}
	sort.Strings(cats)

	builder := strings.Builder{}
	for _, c := range cats {
		builder.WriteString(fmt.Sprintf("## %s\n", c))
		builder.WriteString(dm[c])
	}
	return builder.String()
}

// GenMDDocumentation will generate route documentation in markdown
func (r Route) GenMDDocumentation() string {
	const (
		PermManageOrg   = 1
		PermManageUsers = 2
		PermManageSelf  = 3
	)
	builder := strings.Builder{}
	name := r.Name
	if r.Name == "" {
		name = "UNKNOWN"
	}
	builder.WriteString(fmt.Sprintf("### %s\n", name))
	if r.Description != "" {
		builder.WriteString(fmt.Sprintf("\n%s", r.Description))
	}
	builder.WriteString(fmt.Sprintf("\t%s %s\n", r.Method, r.Path))
	if r.Insecure {
		builder.WriteString("\tToken not required\n")
	}
	if len(r.Permissions) > 0 {
		perms := []string{}
		for _, p := range r.Permissions {
			switch p {
			case PermManageOrg:
				perms = append(perms, fmt.Sprintf("Manage Organization (1)"))
			case PermManageUsers:
				perms = append(perms, fmt.Sprintf("Manage Users (2)"))
			case PermManageSelf:
				perms = append(perms, fmt.Sprintf("Manage Self (3)"))
			}
		}
		builder.WriteString(fmt.Sprintf("\tOrg Permissions: %s\n", strings.Join(perms, ", ")))
	}
	builder.WriteString(fmt.Sprintf("#### Inputs\n"))
	js := fmtJSON(r.Input)
	if js != "" {
		builder.WriteString("```sh\n")
		builder.WriteString(js)
		builder.WriteString("\n```\n")
	} else {
		builder.WriteString("None\n")
	}
	builder.WriteString(fmt.Sprintf("\n#### Outputs\n"))
	js = fmtJSON(r.Output)
	if js != "" {
		builder.WriteString("```sh\n")
		builder.WriteString(js)
		builder.WriteString("\n```\n")
	} else {
		builder.WriteString("None\n\n")
	}
	builder.WriteString("---\n")
	return builder.String()
}

func fmtJSON(val string) string {
	fields := make(map[string]interface{})
	err := json.Unmarshal([]byte(val), &fields)
	if err != nil {
		return val
	}
	js, err := json.MarshalIndent(fields, "", "\t")
	if err != nil {
		return val
	}
	return string(js)
}
