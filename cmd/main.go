package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/akshay0074700747/client-connect/graaph"
	"github.com/akshay0074700747/client-connect/middleware"
	servicediscoveryconsul "github.com/akshay0074700747/client-connect/servicediscovery_consul"
	"github.com/akshay0074700747/proto-files-for-microservices/pb"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
)

func main() {

	userconn, err := servicediscoveryconsul.GetService("user-service")
	if err != nil {
		log.Println(err.Error())
	}

	orderConn, err := servicediscoveryconsul.GetService("order-service")
	if err != nil {
		log.Println(err.Error())
	}

	productConn, err := servicediscoveryconsul.GetService("product-service")
	if err != nil {
		log.Println(err.Error())
	}

	cartConn, err := servicediscoveryconsul.GetService("cart-service")
	if err != nil {
		log.Println(err.Error())
	}
	wishlistConn, err := servicediscoveryconsul.GetService("wishlist-service")
	if err != nil {
		log.Println(err.Error())
	}

	defer func() {
		userconn.Close()
		orderConn.Close()
		productConn.Close()
		cartConn.Close()
		wishlistConn.Close()
	}()

	userRes := pb.NewUserServiceClient(userconn)
	orderRes := pb.NewOrderServiceClient(orderConn)
	productRes := pb.NewProductServiceClient(productConn)
	cartRes := pb.NewCartServiceClient(cartConn)
	wishlistRes := pb.NewWishlistServiceClient(wishlistConn)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err.Error())
	}
	secretString := os.Getenv("SECRET")

	graaph.Initialize(userRes, orderRes, productRes, cartRes, wishlistRes)
	graaph.RetrieveSecret(secretString)
	middleware.InitMiddlewareSecret(secretString)

	// h := handler.New(&handler.Config{
	// 	Schema: &graaph.Schema,
	// 	Pretty: true,
	// })

	// http.Handle("/graphql", h)

	// log.Println("listeninng on port :50001 of api gateway")

	// http.ListenAndServe(":50001", nil)

	h := handler.New(&handler.Config{
		Schema: &graaph.Schema,
		Pretty: true,
	})

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		// Add the http.ResponseWriter to the context.
		ctx := context.WithValue(r.Context(), "httpResponseWriter", w)
		ctx = context.WithValue(ctx, "request", r)

		// Update the request's context.
		r = r.WithContext(ctx)

		// Call the GraphQL handler.
		h.ContextHandler(ctx, w, r)
	})

	log.Println("listening on port :50001 of api gateway")

	tracer, closer := initTracer()

	defer closer.Close()

	graaph.RetrieveTracer(tracer)

	fmt.Println("hi")

	http.ListenAndServe(":50001", nil)

}

func initTracer() (tracer opentracing.Tracer, closer io.Closer) {
	jaegerEndpoint := "http://localhost:14268/api/traces"

	cfg := &config.Configuration{
		ServiceName: "api-gateway",
		Sampler: &config.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: jaegerEndpoint,
		},
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		fmt.Println(err.Error())
	}

	return
}