package graaph

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/akshay0074700747/client-connect/authorize"
	"github.com/akshay0074700747/client-connect/middleware"
	"github.com/akshay0074700747/proto-files-for-microservices/pb"
	"github.com/graphql-go/graphql"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	UsersrvConn  pb.UserServiceClient
	OrdersConn   pb.OrderServiceClient
	ProductsConn pb.ProductServiceClient
	CartConn     pb.CartServiceClient
	WishlistConn pb.WishlistServiceClient
	Secret       []byte
	Tracer       opentracing.Tracer
)

func RetrieveSecret(secretString string) {
	Secret = []byte(secretString)
}

func RetrieveTracer(tr opentracing.Tracer) {
	Tracer = tr
}

func Initialize(usrconn pb.UserServiceClient, ordrconn pb.OrderServiceClient, prodConn pb.ProductServiceClient, cartconn pb.CartServiceClient, wishlistconn pb.WishlistServiceClient) {
	UsersrvConn = usrconn
	OrdersConn = ordrconn
	ProductsConn = prodConn
	CartConn = cartconn
	WishlistConn = wishlistconn
}

var UserType = graphql.NewObject(

	graphql.ObjectConfig{

		Name: "user",

		Fields: graphql.Fields{

			"id": &graphql.Field{
				Type: graphql.Int,
			},

			"email": &graphql.Field{
				Type: graphql.String,
			},

			"password": &graphql.Field{
				Type: graphql.String,
			},

			"mobile": &graphql.Field{
				Type: graphql.String,
			},

			"name": &graphql.Field{
				Type: graphql.String,
			},

			"isAdmin": &graphql.Field{
				Type: graphql.Boolean,
			},

			"isSuAdmin": &graphql.Field{
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

var CartItemType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "cartItem",
		Fields: graphql.Fields{
			"product_id": &graphql.Field{
				Type: graphql.Int,
			},
			"product": &graphql.Field{
				Type: ProductType,
			},
			"quantity": &graphql.Field{
				Type: graphql.Int,
			},
		},
	},
)

var WishlistItemType = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "wishlistItem",
		Fields: graphql.Fields{
			"product_id": &graphql.Field{
				Type: graphql.Int,
			},
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

// var CartType = graphql.NewObject(

// 	graphql.ObjectConfig{

// 		Name: "cart",
// 		Fields: graphql.Fields{

// 		},
// 	},
// )

var RootQuery = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"users": &graphql.Field{
				Type: graphql.NewList(UserType),
				Resolve: middleware.SuAdminOrAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get all users")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					users, err := UsersrvConn.GetAllUsersResponce(ctx, &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					var userss []*pb.UserResponce
					for {
						user, err := users.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
							return nil, err
						}
						userss = append(userss, user)
					}
					return userss, nil
				}),
			},

			"admins": &graphql.Field{
				Type: graphql.NewList(UserType),
				Resolve: middleware.SuAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get all admins")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					admins, err := UsersrvConn.GetAllAdminsResponce(ctx, &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					var adminss []*pb.UserResponce
					for {
						admin, err := admins.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
							return nil, err
						}
						adminss = append(adminss, admin)
					}
					return adminss, nil
				}),
			},

			"user": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.SuAdminOrAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get user")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					return UsersrvConn.GetUser(ctx, &pb.UserRequest{Id: uint32(p.Args["id"].(int))})
				}),
			},
			"admin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.SuAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get admin")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					return UsersrvConn.GetAdmin(ctx, &pb.UserRequest{Id: uint32(p.Args["id"].(int))})
				}),
			},
			"userDetails": &graphql.Field{
				Type: UserType,
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get user details")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)

					user, err := UsersrvConn.GetUser(ctx, &pb.UserRequest{Id: uint32(userIDval)})
					if err != nil {
						fmt.Println(err.Error())
					}

					return user, err
				}),
			},
			"orders": &graphql.Field{
				Type: graphql.NewList(OrderType),
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get all orders")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					orders, err := OrdersConn.GetAllOrdersResponce(ctx, &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					var orderss []*pb.AddOrderResponce
					for {
						order, err := orders.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
							return nil, err
						}
						orderss = append(orderss, order)
					}

					return orderss, nil
				}),
			},
			"order": &graphql.Field{
				Type: graphql.NewList(OrderType),
				// Args: graphql.FieldConfigArgument{
				// 	"user_id": &graphql.ArgumentConfig{
				// 		Type: graphql.NewNonNull(graphql.ID),
				// 	},
				// },
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get order")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)

					user, err := UsersrvConn.GetUser(ctx, &pb.UserRequest{Id: uint32(userIDval)})
					if err != nil {
						return nil, err
					}
					if user.Name == "" {
						return nil, fmt.Errorf("user is not signed up")
					}

					return OrdersConn.GetOrdersByUser(ctx, &pb.GetOrdersByUserRequest{
						UserId: uint32(userIDval),
					})
				}),
			},
			"product": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get product")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					return ProductsConn.GetProduct(ctx, &pb.GetProductByID{
						Id: uint32(p.Args["id"].(int)),
					})
				},
			},
			"products": &graphql.Field{
				Type: graphql.NewList(ProductType),
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get all users")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					products, err := ProductsConn.GetAllProducts(ctx, &emptypb.Empty{})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					var productss []*pb.AddProductResponce
					for {
						product, err := products.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
							return nil, err
						}
						productss = append(productss, product)
					}
					return productss, nil
				},
			},
			"cart": &graphql.Field{
				Type: graphql.NewList(CartItemType),
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get cart")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)

					cart, err := CartConn.GetCart(ctx, &pb.CartRequest{UserId: uint32(userIDval)})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					var cartss []*pb.AddtoCartResponce
					for {
						item, err := cart.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
							return nil, err
						}
						cartss = append(cartss, item)
					}
					return cartss, nil
				}),
			},
			"wishlist": &graphql.Field{
				Type: graphql.NewList(WishlistItemType),
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("get wishlist")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)
					wishlist, err := WishlistConn.GetWishlist(ctx, &pb.WishlistRequest{UserId: uint32(userIDval)})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					var wishh []*pb.AddProductResponce
					for {
						item, err := wishlist.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
							return nil, err
						}
						wishh = append(wishh, item)
					}
					return wishh, nil
				}),
			},
		},
	},
)

var Mutation = graphql.NewObject(
	graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"signup": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"mobile": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("signup user")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					user, err := UsersrvConn.SignupUser(ctx, &pb.SignupUserRequest{
						Name:     p.Args["name"].(string),
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
						Mobile:   p.Args["mobile"].(string),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					_, err = CartConn.CreateCart(ctx, &pb.CartRequest{UserId: user.Id})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					_, err = WishlistConn.CreateWishlist(ctx, &pb.WishlistRequest{UserId: user.Id})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)

					tokenString, err := authorize.GenerateJwt(uint(user.Id), user.IsAdmin, user.IsSuAdmin, Secret)
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: tokenString,
						Path:  "/",
					})

					return user, nil
				},
			},
			"loginUser": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("login user")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					user, err := UsersrvConn.LoginUser(ctx, &pb.LoginRequest{
						Email:     p.Args["email"].(string),
						Password:  p.Args["password"].(string),
						IsAdmin:   false,
						IsSuAdmin: false,
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)

					tokenString, err := authorize.GenerateJwt(uint(user.Id), false, false, Secret)
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: tokenString,
						Path:  "/",
					})

					return user, nil
				},
			},
			"loginAdmin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("login admin")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					user, err := UsersrvConn.LoginUser(ctx, &pb.LoginRequest{
						Email:     p.Args["email"].(string),
						Password:  p.Args["password"].(string),
						IsAdmin:   true,
						IsSuAdmin: false,
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)

					tokenString, err := authorize.GenerateJwt(uint(user.Id), true, false, Secret)
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: tokenString,
						Path:  "/",
					})

					return user, nil
				},
			},
			"loginSuperAdmin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("login suAdmin")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					user, err := UsersrvConn.LoginUser(ctx, &pb.LoginRequest{
						Email:     p.Args["email"].(string),
						Password:  p.Args["password"].(string),
						IsAdmin:   false,
						IsSuAdmin: true,
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					w := p.Context.Value("httpResponseWriter").(http.ResponseWriter)

					tokenString, err := authorize.GenerateJwt(uint(user.Id), false, true, Secret)
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					http.SetCookie(w, &http.Cookie{
						Name:  "jwtToken",
						Value: tokenString,
						Path:  "/",
					})

					return user, nil
				},
			},
			"addAdmin": &graphql.Field{
				Type: UserType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"mobile": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
				},
				Resolve: middleware.SuAdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("add admin")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					admin, err := UsersrvConn.AddAdmin(ctx, &pb.SignupUserRequest{
						Name:     p.Args["name"].(string),
						Email:    p.Args["email"].(string),
						Password: p.Args["password"].(string),
						Mobile:   p.Args["mobile"].(string),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return admin, nil
				}),
			},
			"checkoutCart": &graphql.Field{
				Type: OrderType,
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("checkout cart")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)

					user, err := UsersrvConn.GetUser(ctx, &pb.UserRequest{Id: uint32(userIDval)})
					if err != nil {
						return nil, err
					}
					if user.IsAdmin {
						return nil, fmt.Errorf("Admin cannot order")
					}

					cart, err := CartConn.GetCart(ctx, &pb.CartRequest{UserId: uint32(userIDval)})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					var products []*pb.ProductwithQuantity
					for {
						item, err := cart.Recv()
						if err == io.EOF {
							break
						}
						if err != nil {
							fmt.Println(err)
							return nil, err
						}
						products = append(products, &pb.ProductwithQuantity{
							Id:       item.Product.Id,
							Quantity: uint32(item.Quantity),
						})
					}

					order, err := OrdersConn.AddOrder(ctx, &pb.AddOrderRequest{
						UserId:   uint32(userIDval),
						Products: products,
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					_, err = CartConn.TruncateCart(ctx, &pb.CartRequest{UserId: uint32(userIDval)})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}

					return order, nil
				}),
			},
			"AddProduct": &graphql.Field{
				Type: ProductType,
				Args: graphql.FieldConfigArgument{
					"name": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.String),
					},
					"price": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
					"stock": &graphql.ArgumentConfig{
						Type: graphql.NewNonNull(graphql.Int),
					},
				},
				Resolve: middleware.AdminMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("add product")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					fmt.Println("here reached...")
					products, err := ProductsConn.AddProducts(ctx, &pb.AddProductRequest{
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

					span := Tracer.StartSpan("update stock")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					return ProductsConn.UpdateStock(ctx, &pb.UpdateStockRequest{
						Id:       p.Args["id"].(uint32),
						Stock:    p.Args["stock"].(int32),
						Increase: p.Args["increase"].(bool),
					})
				}),
			},
			"addToCart": &graphql.Field{
				Type: CartItemType,
				Args: graphql.FieldConfigArgument{
					"product_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"quantity": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("add to cart")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)

					cart, err := CartConn.AddtoCart(ctx, &pb.AddtoCartRequest{
						UserId:    uint32(userIDval),
						ProductId: uint32(p.Args["product_id"].(int)),
						Quantity:  int32(p.Args["quantity"].(int)),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return cart, nil
				}),
			},
			"removeFromCart": &graphql.Field{
				Type: CartItemType,
				Args: graphql.FieldConfigArgument{
					"product_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("remove from cart")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)
					cart, err := CartConn.DeleteCartItem(ctx, &pb.AddtoCartRequest{
						UserId:    uint32(userIDval),
						ProductId: uint32(p.Args["product_id"].(int)),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return cart, nil
				}),
			},
			"updateCartItemQty": &graphql.Field{
				Type: CartItemType,
				Args: graphql.FieldConfigArgument{
					"product_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"quantity": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"isIncreasing": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
				},
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("update cart item qty")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)
					cart, err := CartConn.ChangeQty(ctx, &pb.ChangeQtyRequest{
						UserId:     uint32(userIDval),
						ProductId:  uint32(p.Args["product_id"].(int)),
						Quantity:   int32(p.Args["quantity"].(int)),
						IsIncrease: p.Args["isIncreasing"].(bool),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return cart, nil
				}),
			},
			"addToWishlist": &graphql.Field{
				Type: WishlistItemType,
				Args: graphql.FieldConfigArgument{
					"product_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("add to wishlist")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)
					wishlist, err := WishlistConn.AddtoWishlist(ctx, &pb.AddtoWishlistRequest{
						UserId:    uint32(userIDval),
						ProductId: uint32(p.Args["product_id"].(int)),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return wishlist, nil
				}),
			},
			"removeWishlistItem": &graphql.Field{
				Type: WishlistItemType,
				Args: graphql.FieldConfigArgument{
					"product_id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("remove wishlist item")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)
					wishlist, err := WishlistConn.DeleteWishlistItem(ctx, &pb.AddtoWishlistRequest{
						UserId:    uint32(userIDval),
						ProductId: uint32(p.Args["product_id"].(int)),
					})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return wishlist, nil
				}),
			},
			"transferToCart": &graphql.Field{
				Type: CartItemType,
				Resolve: middleware.ClientMiddleware(func(p graphql.ResolveParams) (interface{}, error) {

					span := Tracer.StartSpan("transfer to cart")
					defer span.Finish()
					ctx := opentracing.ContextWithSpan(context.Background(), span)

					userIDval := p.Context.Value("userID").(uint)
					cart, err := CartConn.TrasferWishlist(ctx, &pb.CartRequest{UserId: uint32(userIDval)})
					if err != nil {
						fmt.Println(err.Error())
						return nil, err
					}
					return cart, nil
				}),
			},
		},
	},
)

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:    RootQuery,
	Mutation: Mutation,
})
