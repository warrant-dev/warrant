package grpcClients

import (
	"github.com/rs/zerolog/log"
	"github.com/warrant-dev/warrant/pkg/config"
	"github.com/warrant-dev/warrant/pkg/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 设置最大接收消息大小为100MB
var maxReceiveMessageSize = 100 * 1024 * 1024

// 设置最大发送消息大小为100MB
var maxSendMessageSize = 100 * 1024 * 1024

var GroupServiceClient pb.GroupServiceClient

var InternalOrgServiceClient pb.InternalOrgServiceClient

func Start(config config.WarrantConfig) {
	mainServerConn, err := grpc.NewClient(config.Grpc.Client.MainServerHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxReceiveMessageSize),
			grpc.MaxCallSendMsgSize(maxSendMessageSize),
		))
	if err != nil {
		log.Fatal().Err(err).Msg("init: could not init grpc connect for main-server: " + config.Grpc.Client.MainServerHost)
	}

	GroupServiceClient = pb.NewGroupServiceClient(mainServerConn)
	InternalOrgServiceClient = pb.NewInternalOrgServiceClient(mainServerConn)
}
