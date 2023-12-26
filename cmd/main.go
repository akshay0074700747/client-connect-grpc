package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/akshay0074700747/client-connect/graaph"
	"github.com/akshay0074700747/client-connect/middleware"
	"github.com/akshay0074700747/proto-files-for-microservices/pb"
	"github.com/graphql-go/handler"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {

	userconn, err := grpc.Dial(":50002", grpc.WithInsecure())
	if err != nil {
		log.Println(err.Error())
	}

	orderConn, err := grpc.Dial(":50003", grpc.WithInsecure())
	if err != nil {
		log.Println(err.Error())
	}

	productConn, err := grpc.Dial(":50004", grpc.WithInsecure())
	if err != nil {
		log.Println(err.Error())
	}

	defer func() {
		userconn.Close()
		orderConn.Close()
		productConn.Close()
	}()

	userRes := pb.NewUserServiceClient(userconn)
	orderRes := pb.NewOrderServiceClient(orderConn)
	productRes := pb.NewProductServiceClient(productConn)

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal(err.Error())
	}
	secretString := os.Getenv("SECRET")

	graaph.Initialize(userRes, orderRes, productRes)
	graaph.RetrieveSecret(secretString)
	middleware.InitializeMiddleware(userRes)
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

	http.ListenAndServe(":50001", nil)

}
