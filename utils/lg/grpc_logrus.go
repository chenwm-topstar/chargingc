package lg

// import (
// 	"time"
// 	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
// 	"github.com/sirupsen/logrus"
// 	"google.golang.org/grpc"
// )

// // UnaryServerInterceptor injects log context into outgoing grpc request.
// func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
// 	opts := []grpc_logrus.Option{
// 		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
// 		grpc_logrus.WithTimestampFormat(time.RFC3339),
// 	}

// 	return grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logrus.New()), opts...)
// }

// // StreamServerInterceptor retrieve log context from incoming grpc request.
// func StreamServerInterceptor() grpc.StreamServerInterceptor {
// 	opts := []grpc_logrus.Option{
// 		grpc_logrus.WithLevels(grpc_logrus.DefaultCodeToLevel),
// 		grpc_logrus.WithTimestampFormat(time.RFC3339),
// 	}
// 	return grpc_logrus.StreamServerInterceptor(logrus.NewEntry(logrus.New()), opts...)
// }
