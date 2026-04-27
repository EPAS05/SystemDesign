package router

import (
	"classifier/internal/api_handlers"

	"github.com/gorilla/mux"
)

type Handlers struct {
	Node      *api_handlers.NodeHandler
	Unit      *api_handlers.UnitHandler
	Enum      *api_handlers.EnumHandler
	Product   *api_handlers.ProductHandler
	Customer  *api_handlers.CustomerHandler
	Parameter *api_handlers.ParameterHandler
}

func New(handlers Handlers) *mux.Router {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()

	registerNodeRoutes(api, handlers.Node)
	registerUnitRoutes(api, handlers.Unit)
	registerEnumRoutes(api, handlers.Enum)
	registerProductRoutes(api, handlers.Product)
	registerParameterRoutes(api, handlers.Parameter)
	registerCustomerRoutes(api, handlers.Customer)

	return r
}
