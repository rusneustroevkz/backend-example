package http

import (
	"github.com/rusneustroevkz/http-server/internal/metrics"
	"time"

	_ "github.com/rusneustroevkz/http-server/docs"
	"github.com/rusneustroevkz/http-server/internal/config"
	"github.com/rusneustroevkz/http-server/internal/graph/generated"
	"github.com/rusneustroevkz/http-server/internal/graph/resolvers"
	"github.com/rusneustroevkz/http-server/pkg/logger"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Router struct {
	cfg      *config.Config
	resolver *resolvers.Resolver
	log      logger.Logger
}

func NewRouter(cfg *config.Config, resolver *resolvers.Resolver, log logger.Logger) *Router {
	return &Router{
		cfg:      cfg,
		resolver: resolver,
		log:      log,
	}
}

func (r *Router) Mount(routes ...Route) *chi.Mux {
	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	if r.cfg.App.RequestLogEnabled {
		mux.Use(r.log.RequestLogger(r.log))
	}
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(time.Second * time.Duration(r.cfg.PublicServer.Timeout)))
	mux.Use(metrics.RequestMetrics)

	for _, route := range routes {
		mux.Mount(route.Pattern(), route.Routes())
	}

	if !r.cfg.App.Production {
		mux.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("doc.json")))
		mux.Handle("/graph/playground", playground.Handler("GraphQL playground", "/graph/query"))
	}

	schemaConfig := generated.Config{Resolvers: r.resolver}
	schema := generated.NewExecutableSchema(schemaConfig)
	srv := handler.New(schema)
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	mux.Handle("/graph/query", srv)

	return mux
}

type Route interface {
	Routes() *chi.Mux

	Pattern() string
}
