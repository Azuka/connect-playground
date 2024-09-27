package server

import (
	"connectrpc.com/connect"
	"context"
	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"net/http"
	"time"

	otelconnect "connectrpc.com/otelconnect"
	"connectrpc.com/validate"

	connectcors "connectrpc.com/cors"
	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/riandyrn/otelchi"
	"go.azuka.me/connect-playground/internal/app/todos"
	"go.azuka.me/connect-playground/internal/pkg/proto/playground/v1/v1connect"
	"go.opentelemetry.io/otel"
)

func hooks(logger *otelzap.Logger) []connect.HandlerOption {

	telemetryInterceptor, err := otelconnect.NewInterceptor()

	if err != nil {
		logger.Error("Invalid connect interceptor setup", zap.Error(errors.WithStack(err)))
	}

	validateInterceptor, err := validate.NewInterceptor()
	if err != nil {
		logger.Error("Invalid connect interceptor setup", zap.Error(errors.WithStack(err)))
	}

	return []connect.HandlerOption{
		connect.WithInterceptors(
			telemetryInterceptor,
			validateInterceptor,
		),
		connect.WithRecover(
			func(ctx context.Context, spec connect.Spec, header http.Header, a any) error {
				logger.ErrorContext(
					ctx,
					"panic occurred in handler",
					zap.Error(err),
					zap.String("procedure", spec.Procedure),
				)
				return connect.NewError(connect.CodeFailedPrecondition, err)
			},
		),
	}
}

func Handler(logger *otelzap.Logger) (http.Handler, error) {
	router := chi.NewRouter()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	_, span := otel.Tracer("server").Start(ctx, "bootstrap")
	defer span.End()

	router.Use(
		otelchi.Middleware("server",
			otelchi.WithChiRoutes(router),
			otelchi.WithRequestMethodInSpanName(true),
		),
	)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Handle grpc-defined routes
	handleRoute(
		router,
		handler(v1connect.NewTodoServiceHandler(todos.New(logger), hooks(logger)...)),
	)

	router.Get("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	})

	router.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
		_, _ = writer.Write([]byte("Unable to handle request " + request.URL.String())) //nolint
	})

	return router, nil
}

type handlerObject struct {
	handler http.Handler
	prefix  string
}

func handler(prefix string, handler http.Handler) handlerObject {
	return handlerObject{
		handler: handler,
		prefix:  prefix,
	}
}

func handleRoute(router *chi.Mux, h handlerObject) {
	formatted := h.prefix + "*"
	router.Handle(formatted, h.handler)
}
