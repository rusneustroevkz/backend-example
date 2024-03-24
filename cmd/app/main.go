package main

import (
	"context"

	"github.com/rusneustroevkz/http-server/internal/config"
	"github.com/rusneustroevkz/http-server/internal/graph/resolvers"
	kafkaClient "github.com/rusneustroevkz/http-server/internal/kafka"
	grpcServer "github.com/rusneustroevkz/http-server/internal/server/grpc"
	httpServer "github.com/rusneustroevkz/http-server/internal/server/http"
	"github.com/rusneustroevkz/http-server/pkg/logger"
	categoriesGRPCHandlers "github.com/rusneustroevkz/http-server/src/categories/handlers/grpc"
	productGraph "github.com/rusneustroevkz/http-server/src/product/handlers/graph"
	productGRPCHandlers "github.com/rusneustroevkz/http-server/src/product/handlers/grpc"
	productsRest "github.com/rusneustroevkz/http-server/src/product/handlers/rest"

	"github.com/go-chi/chi/v5"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Store server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
func main() {
	fx.New(
		fx.WithLogger(func(log logger.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.Logger()}
		}),
		fx.Provide(
			config.NewConfig,
			logger.NewLogger,
			httpServer.NewHTTPServer,
			productsRest.NewProductsRest,
			productGRPCHandlers.NewProductsGRPCServer,
			categoriesGRPCHandlers.NewCategoriesGRPCServer,
			resolvers.NewResolver,
			productGraph.NewProductGraph,
			kafkaClient.NewClient,
			func(
				cfg *config.Config,
				resolver *resolvers.Resolver,
				productRest *productsRest.ProductsRest,
			) *chi.Mux {
				return httpServer.Routes(cfg, resolver, productRest)
			},
			func(
				cfg *config.Config,
				log logger.Logger,
				productsGRPCServer *productGRPCHandlers.ProductsGRPCServer,
				categoriesGRPCServer *categoriesGRPCHandlers.CategoriesGRPCServer,
			) *grpcServer.Server {
				return grpcServer.NewGRPCServer(
					cfg,
					log,
					productsGRPCServer,
					categoriesGRPCServer,
				)
			},
		),
		fx.Invoke(
			func(lc fx.Lifecycle, srv *httpServer.Server) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return srv.Start(ctx)
					},
					OnStop: func(ctx context.Context) error {
						return srv.Stop(ctx)
					},
				})
			},
			func(lc fx.Lifecycle, srv *grpcServer.Server, petsGRPCServer *petsGRPCHandlers.PetsGRPCServer) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						srv.MountServices(petsGRPCServer)

						return srv.Start(ctx)
					},
					OnStop: func(ctx context.Context) error {
						return srv.Stop(ctx)
					},
				})
			},
			func(lc fx.Lifecycle, client *kafkaClient.Client) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						return client.Run(ctx)
					},
					OnStop: func(ctx context.Context) error {
						return client.Stop(ctx)
					},
				})
			},
		),
	).Run()
}
