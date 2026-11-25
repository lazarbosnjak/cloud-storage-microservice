package grpc

import (
	"context"
	"log/slog"
	"time"

	getbyuser "notification-service/internal/application/notification/get_by_user"
	"notification-service/internal/domain/model"

	pb "proto/notification"

	"github.com/gocql/gocql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedNotificationServer

	NotificationGetByUserHandler *getbyuser.Handler
}

func NewServer(
	notificationGetByUserHandler *getbyuser.Handler,
) *Server {
	return &Server{
		NotificationGetByUserHandler: notificationGetByUserHandler,
	}
}

func (s *Server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingReply, error) {
	time.Sleep(time.Second * 5)

	return &pb.PingReply{Message: "hello from notification-service"}, nil
}

func (s *Server) GetAllByUser(
	ctx context.Context,
	in *pb.GetAllByUserRequest,
) (*pb.GetAllByUserResponse, error) {
	userID, err := gocql.ParseUUID(in.UserId)
	if err != nil {
		slog.Error("error parsing uuid", "err", err)
		return nil, status.Error(codes.InvalidArgument, "error parsing uuid")
	}

	query := getbyuser.Query{
		UserID: userID,
	}

	res, err := s.NotificationGetByUserHandler.Handle(ctx, query)
	if err != nil {
		slog.Error("error fetching notifications", "err", err)
		return nil, status.Error(codes.Internal, "error fetching notifications")
	}

	items := make([]*pb.GetAllByUserResponseItem, 0, len(res.Items))
	for _, resItem := range res.Items {
		item, ok := toProtoItem(resItem)
		if !ok {
			slog.Warn("skipping notification with unknown payload",
				"notification_id", resItem.NotificationID, "event_type", resItem.EventType)
			continue
		}
		items = append(items, item)
	}

	return &pb.GetAllByUserResponse{
		Items: items,
	}, nil
}

func toProtoItem(n *getbyuser.ResponseItem) (*pb.GetAllByUserResponseItem, bool) {
	item := &pb.GetAllByUserResponseItem{
		NotificationId: n.NotificationID.String(),
		UserId:         n.UserID.String(),
		EventType:      string(n.EventType),
	}

	switch v := n.Payload.(type) {
	case model.FileShared:
		item.Payload = &pb.GetAllByUserResponseItem_FileShared{
			FileShared: &pb.FileShared{
				FileId:    v.FileID,
				FileName:  v.FileName,
				ActorId:   v.ActorID,
				ActorName: v.ActorName,
			},
		}
	case model.FolderShared:
		item.Payload = &pb.GetAllByUserResponseItem_FolderShared{
			FolderShared: &pb.FolderShared{
				FolderId:   v.FolderID,
				FolderName: v.FolderName,
				ActorId:    v.ActorID,
				ActorName:  v.ActorName,
			},
		}
	case model.FileAddedToSharedFolder:
		item.Payload = &pb.GetAllByUserResponseItem_FileAddedToSharedFolder{
			FileAddedToSharedFolder: &pb.FileAddedToSharedFolder{
				FileId:           v.FileID,
				FileName:         v.FileName,
				SharedFolderId:   v.SharedFolderID,
				SharedFolderName: v.SharedFolderName,
				ActorId:          v.ActorID,
				ActorName:        v.ActorName,
			},
		}
	case model.SharedFileNewVersion:
		item.Payload = &pb.GetAllByUserResponseItem_SharedFileNewVersion{
			SharedFileNewVersion: &pb.SharedFileNewVersion{
				FileId:    v.FileID,
				FileName:  v.FileName,
				ActorId:   v.ActorID,
				ActorName: v.ActorName,
			},
		}
	default:
		return nil, false
	}
	return item, true
}
