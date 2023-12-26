package graaph

import (
	"context"
	"fmt"
	"net/http"

	"github.com/akshay0074700747/client-connect/authorize"
	"github.com/akshay0074700747/client-connect/middleware"
	"github.com/akshay0074700747/proto-files-for-microservices/pb"
	"github.com/graphql-go/graphql"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	UsersrvConn  pb.UserServiceClient
	OrdersConn   pb.OrderServiceClient
	ProductsConn pb.ProductServiceClient
	Secret       []byte
)

func RetrieveSecret(secretString string) {
	Secret = []byte(secretString)
}

func Initialize(usrconn pb.UserServiceClient, ordrconn pb.OrderServiceClient, prodConn pb.ProductServiceClient) {
	UsersrvConn = usrconn
	OrdersConn = ordrconn
	ProductsConn = prodConn
}

var UserType = graphql.NewObject(

	graphql.ObjectConfig{

		Name: "user",

		Fields: graphql.Fields{

			"id": &graphql.Field{
				Type: graphql.Int,
			},

			"name": &graphql.Field{
				Type: graphql.String,
			},

			"isAdmin": &graphql.Field{
				Type: graphql.Boolean,
			},
		},
	},
)

var ProductType = graphql.NewObject(

	graphql.ObjectConfig{

		Name: "product",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"price": &graphql.Field{
				Type: graphql.Int,
			},
			"stock": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)

var OrderType = graphql.NewObject(

	graphql.ObjectConfig{

		Name: "order",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"user": &graphql.Field{
				Type: UserType,
			},
			"price": &graphql.Field{
				Type: graphql.Int,
			},
			"products": &graphql.Field{
				Type: graphql.NewList(ProductType),
			},
		},
	},
)

var RootQuery = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"users": &graphql.Field{
				Type: graphql.NewList(UserType),
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					users, err := UsersrvConn.GetAllUsersResponce(context.Background(), &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
					}
					return users.Users, nil
				}),
			},

			"user": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					return UsersrvConn.GetUser(context.Background(), &pb.UserRequest{Id: uint32(p.Args["id"].(int))})
				}),
			},
			"userDetails": &graphql.Field{
				Type: UserType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					r := p.Context.Value("request").(*http.Request)
					cookie, err := r.Cookie("jwtToken")
					if err != nil {
						return nil, err
					}
					if cookie == nil {
						return nil, fmt.Errorf("you are not logged in")
					}

					token := cookie.Value
					auth, err := authorize.ValidateToken(token, Secret)
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					userIDval := auth["userID"].(uint)

					user, err := UsersrvConn.GetUser(context.Background(), &pb.UserRequest{Id: uint32(userIDval)})
					if err != nil {
						return nil, err
					}

					return user, err
				},
			},
			"orders": &graphql.Field{
				Type: graphql.NewList(OrderType),
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {
					orders, err := OrdersConn.GetAllOrdersResponce(context.Background(), &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
					}

					return orders.Orders, err
				}),
			},
			"order": &graphql.Field{
				Type: graphql.NewList(OrderType),
				// Args: graphql.FieldConfigArgument{
				// 	"user_id": &graphql.ArgumentConfig{
				// 		Type: graphql.NewNonNull(graphql.ID),
				// 	},
				// },
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					r := p.Context.Value("request").(*http.Request)
					cookie, err := r.Cookie("jwtToken")
					if err != nil {
						return nil, err
					}
					if cookie == nil {
						return nil, fmt.Errorf("you are not logged in")
					}

					token := cookie.Value
					auth, err := authorize.ValidateToken(token, Secret)
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					userIDval := auth["userID"].(uint)

					user, err := UsersrvConn.GetUser(context.Background(), &pb.UserRequest{Id: uint32(userIDval)})
					if err != nil {
						return nil, err
					}
					if user.Name == "" {
						return nil, fmt.Errorf("user is not signed up")
					}

					return OrdersConn.GetOrdersByUser(context.Background(), &pb.GetOrdersByUserRequest{
						UserId: uint32(userIDval),
					})
				},
			},
			"product": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					return ProductsConn.GetProduct(context.Background(), &pb.GetProductByID{
						Id: uint32(p.Args["id"].(int)),
					})
				},
			},
			"products": &graphql.Field{
				Type: graphql.NewList(ProductType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					products, err := ProductsConn.GetAllProducts(context.Background(), &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
					}
					return products.Products, err
				},
			},
		},
	},
)

var Mutation = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"addUser": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"isAdmin": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Boolean),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					user, err := UsersrvConn.AddUser(context.Background(), &pb.AddUserRequest{
						Name:    p.Args["name"].(string),
						IsAdmin: p.Args["isAdmin"].(bool),
					})
					if err != nil {
						return nil, err
					}

					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)

					tokenString, err := authorize.GenerateJwt(uint(user.Id), user.IsAdmin, Secret)
					if err != nil {
						return nil, err
					}

					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: tokenString,
						Path:  "/",
					})

					fmt.Println(user.Id)
					fmt.Println(user.Name)
					return user, nil
				},
			},
			"addOrder": &graphql.Field{
				Type: OrderType,
				Args: graphql.FieldConfigArgument{
					"product_ids": &graphql.ArgumentConfig{
						Type: graphql.NewList(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					r := p.Context.Value("request").(*http.Request)
					cookie, err := r.Cookie("jwtToken")
					if err != nil {
						return nil, err
					}
					if cookie == nil {
						return nil, fmt.Errorf("you are not logged in")
					}

					token := cookie.Value
					auth, err := authorize.ValidateToken(token, Secret)
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					userIDval := auth["userID"].(uint)

					user, err := UsersrvConn.GetUser(context.Background(), &pb.UserRequest{Id: uint32(userIDval)})
					if err != nil {
						return nil, err
					}
					if user.IsAdmin {
						return nil, fmt.Errorf("Admin cannot order")
					}

					prodids := p.Args["product_ids"].([]interface{})
					var productids []int32
					for _, prod := range prodids {
						productids = append(productids, int32(prod.(int)))
					}
					return OrdersConn.AddOrder(context.Background(), &pb.AddOrderRequest{
						UserId:     uint32(userIDval),
						ProductIDs: productids,
					})
				},
			},
			"AddProduct": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"price": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"stock": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					fmt.Println("here reached...")
					products, err := ProductsConn.AddProducts(context.Background(), &pb.AddProductRequest{
						Name:  p.Args["name"].(string),
						Price: int32(p.Args["price"].(int)),
						Stock: int32(p.Args["stock"].(int)),
					})
					if err != nil {
						fmt.Println(err.Error())
					}
					return products, nil
				}),
			},
			"updateStock": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.ID),
					},
					"stock": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"increase": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Boolean),
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					return ProductsConn.UpdateStock(context.Background(), &pb.UpdateStockRequest{
						Id:       p.Args["id"].(uint32),
						Stock:    p.Args["stock"].(int32),
						Increase: p.Args["increase"].(bool),
					})
				}),
			},
		},
	},
)

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    RootQuery,
	Mutation: Mutation,
})

/*
Now i have to implement login and logout
i have to check wheather the increment and decrement stock works and also to chek the orders of the user

*/
