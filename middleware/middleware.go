package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/akshay0074700747/client-connect/authorize"
	"github.com/akshay0074700747/proto-files-for-microservices/pb"
	"github.com/graphql-go/graphql"
)

var (
	usersrvConn pb.UserServiceClient
	secret      []byte
)

func InitMiddlewareSecret(secretString string) {
	secret = []byte(secretString)
}

func InitializeMiddleware(usrconn pb.UserServiceClient) {
	usersrvConn = usrconn
}

func ClientMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {

		r := p.Context.Value("request").(*http.Request)
		cookie, err := r.Cookie("jwtToken")
		if err != nil {
			return nil, err
		}
		if cookie == nil {
			return nil, fmt.Errorf("you are not logged in")
		}

		token := cookie.Value
		auth, err := authorize.ValidateToken(token, secret)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		userID := auth["userID"].(string)
		userIDval, err := strconv.ParseUint(userID, 10, 64)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		user, err := usersrvConn.GetUser(context.Background(), &pb.UserRequest{Id: uint32(userIDval)})
		if err != nil {
			return nil, err
		}
		if user.Name == "" {
			return nil, fmt.Errorf("user is not signed up")
		}
		return next(p)
	}
}

func AdminMiddleware(next graphql.FieldResolveFn) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {

		r := p.Context.Value("request").(*http.Request)
		cookie, err := r.Cookie("jwtToken")
		if err != nil {
			return nil, err
		}
		if cookie == nil {
			return nil, fmt.Errorf("you are not logged in")
		}

		token := cookie.Value
		auth, err := authorize.ValidateToken(token, secret)
		if err != nil {
			fmt.Println(err.Error())
			return nil, err
		}

		userIDval := auth["userID"].(uint)

		user, err := usersrvConn.GetUser(context.Background(), &pb.UserRequest{Id: uint32(userIDval)})
		if err != nil {
			return nil, err
		}
		if user.Name == "" {
			return nil, fmt.Errorf("user is not signed up")
		}
		if !user.IsAdmin {
			return nil, fmt.Errorf("you are not an admin to perform this action")
		}
		return next(p)
	}
}
