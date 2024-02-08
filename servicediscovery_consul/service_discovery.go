package servicediscoveryconsul

import (
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
)

func GetService(service string) (*grpc.ClientConn, error) {

	client, err := consulapi.NewClient(consulapi.DefaultConfig())
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("kkkkkkkkkkkkkkkkkkkkkkkkkkkkk")
		return nil, err
	}

	services, err := client.Agent().Services()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	Conn, err := grpc.Dial(fmt.Sprintf("%s:%d", services[service].Address, services[service].Port), grpc.WithInsecure())
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("iiiiiiiiiiiiiiiiiiiiiiiiiiiiiiii")
		return nil, err
	}

	return Conn, nil

}
