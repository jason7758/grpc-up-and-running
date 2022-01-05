package service

import (
	"context"
	"fmt"
	wrapper "github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	pb "ordermgt/service/ecommerce"
	"strings"
)

const (
	port           = ":50051"
	orderBatchSize = 3
)

var orderMap = make(map[string]pb.Order)

type server struct {
	oderMap map[string]pb.Order
}

//
// AddOrder
//  @Description: simple RPC
//  @receiver s
//  @param ctx
//  @param orderReq
//  @return *wrapper.StringValue
//  @return error
//
func (s *server) AddOrder(ctx context.Context, orderReq *pb.Order) (*wrapper.StringValue, error) {
	log.Printf("[DEBUG] Adding order. ID: %v", orderReq.Id)
	orderMap[orderReq.Id] = *orderReq
	return &wrapper.StringValue{Value: "Order added: " + orderReq.Id}, nil
}

func (s *server) GetOrder(ctx context.Context, orderId *wrapper.StringValue) (*pb.Order, error) {
	ord, existes := orderMap[orderId.Value]
	if existes {
		return &ord, status.New(codes.OK, "").Err()
	}
	return nil, status.Errorf(codes.NotFound, "Order does not exist. :", orderId)
}
func (s *server) SearchOrders(searchQuery *wrapper.StringValue, stream pb.OrderManagement_SearchOrdersServer) error {
	for key, order := range orderMap {
		log.Print(key, order)
		for _, itemStr := range order.Items {
			log.Print(itemStr)
			if strings.Contains(itemStr, searchQuery.Value) {
				err := stream.Send(&order)
				if err != nil {
					return fmt.Errorf("error sending message to stream : %v", err)
				}
				log.Print("Matching orde found: " + key)
				break
			}
		}
	}
	return nil
}

func (s *server) UpdateOrder(stream pb.OrderManagement_UpdateOrderServer) error {
	orderStr := "Update Order IDs :"
	for {
		order, err := stream.Recv()
		if err == io.EOF {
			// Finished reading the order stream
			return stream.SendAndClose(&wrapper.StringValue{Value: "Orders processed " + orderStr})
		}
		if err != nil {
			return err
		}
		// Update order
		orderMap[order.Id] = *order
		log.Printf("Orders ID:%s - %s", order.Id, "Update")
		orderStr += order.Id + ", "
	}
}

func (s *server) ProcessOrders(stream pb.OrderManagement_ProcessOrderServer) error {

}
