package lg

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"cchome-admin-topstar/utils/internal/shared"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	msgMetadataKey    = "lc-msg"
	keysMetadataKey   = "lc-keys"
	valuesMetadataKey = "lc-values"
	uuidMetadataKey   = "lc-uuid"
	fromMetadataKey   = "lc-from"
	sinceMetadataKey  = "lc-since"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	trafficCounter = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "chargingc_service_handling",
	}, []string{"from_service", "from_version", "to_service", "to_version", "method"})
	streamTrafficCounter = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "chargingc_service_stream_handling",
	}, []string{"from_service", "from_version", "to_service", "to_version", "method"})
)

func appendLogContext(ctx context.Context, lc *LogContext, from string) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	if lc != nil {
		// Override the log context value in metadata.
		md.Set(msgMetadataKey, lc.msg...)
		md.Set(keysMetadataKey, lc.keys...)
		md.Set(valuesMetadataKey, lc.values...)
	}
	if lc == nil || lc.uuid == 0 {
		md.Set(uuidMetadataKey, strconv.FormatInt(rand.Int63(), 10))
	} else {
		md.Set(uuidMetadataKey, strconv.FormatInt(lc.uuid, 10))
	}
	if lc == nil || lc.since.IsZero() {
		md.Set(sinceMetadataKey, strconv.FormatInt(time.Now().UnixNano(), 10))
	} else {
		md.Set(sinceMetadataKey, strconv.FormatInt(lc.since.UnixNano(), 10))
	}
	md.Set(fromMetadataKey, from)
	return metadata.NewOutgoingContext(ctx, md)
}

func extractLogContext(ctx context.Context) *LogContext {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	lc := &LogContext{}
	lc.msg = md.Get(msgMetadataKey)
	lc.keys = md.Get(keysMetadataKey)
	lc.values = md.Get(valuesMetadataKey)
	fromSlice := md.Get(fromMetadataKey)
	if len(fromSlice) > 0 {
		lc.from = fromSlice[0]
	}
	uuidSlice := md.Get(uuidMetadataKey)
	if len(uuidSlice) > 0 {
		lc.uuid, _ = strconv.ParseInt(uuidSlice[0], 10, 64)
	}
	sinceSlice := md.Get(sinceMetadataKey)
	if len(sinceSlice) > 0 {
		i, _ := strconv.ParseInt(sinceSlice[0], 10, 64)
		lc.since = time.Unix(0, i)
	}
	if lc.Empty() {
		return nil
	}
	return lc
}

// UnaryClientInterceptor injects log context into outgoing grpc request.
func UnaryClientInterceptor(from string) func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = appendLogContext(ctx, FromContext(ctx), from)
		if !IsDebugging() {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		td := TimeFuncDuration()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			Debugc(ctx, "Failed to invoke method %s invoke_time=%s invoke_err=%s", method, td(), err)
		} else {
			Debugc(ctx, "Succeed to invoke method %s invoke_time=%s", method, td())
		}
		return err
	}
}

// StreamClientInterceptor injects log context into outgoing grpc request.
func StreamClientInterceptor(from string) func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = appendLogContext(ctx, FromContext(ctx), from)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// UnaryServerInterceptor retrieve log context from incoming grpc request.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	lc := extractLogContext(ctx)
	if lc != nil {
		ctx = context.WithValue(ctx, logContextKey, lc)
	}
	td := TimeFuncDuration()

	prefix := fmt.Sprintf("[%s]", strings.TrimPrefix(info.FullMethod, "/"))
	ctx = With(ctx, prefix)
	ret, err := handler(ctx, req)
	duration := td()
	if err != nil {
		Debugc(ctx, "Failed to handle method %s handle_time=%s handle_err=%s", info.FullMethod, duration, err)
	} else {
		Debugc(ctx, "Succeed to handle method %s handle_time=%s", info.FullMethod, duration)
	}
	fromService, fromVersion := splitServiceVersion(lc.From())
	toService, toVersion := splitServiceVersion(shared.GetServiceName())
	trafficCounter.WithLabelValues(fromService, fromVersion, toService, toVersion, info.FullMethod).Observe(duration.Seconds())
	return ret, err
}

func splitServiceVersion(str string) (serviceName string, version string) {
	segs := strings.SplitN(str, ":", 2)
	serviceName = segs[0]
	if len(segs) > 1 {
		version = segs[1]
	}
	return
}

// StreamServerInterceptor retrieve log context from incoming grpc request.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	lc := extractLogContext(ctx)
	if lc != nil {
		ctx = context.WithValue(ctx, logContextKey, lc)
	}
	prefix := fmt.Sprintf("[%s]", strings.TrimPrefix(info.FullMethod, "/"))
	ctx = With(ctx, prefix)
	ss = newInjectServerStream(ctx, ss)
	td := TimeFuncDuration()
	// Wrap a server stream implementation to modify the context to include the data.
	err := handler(srv, ss)
	duration := td()

	if lc != nil {
		fromService, fromVersion := splitServiceVersion(lc.From())
		toService, toVersion := splitServiceVersion(shared.GetServiceName())
		streamTrafficCounter.WithLabelValues(fromService, fromVersion, toService, toVersion, info.FullMethod).Observe(duration.Seconds())
	}
	return err
}

type injectServerStream struct {
	ctx context.Context
	ss  grpc.ServerStream
}

func newInjectServerStream(ctx context.Context, ss grpc.ServerStream) *injectServerStream {
	return &injectServerStream{
		ss:  ss,
		ctx: ctx,
	}
}

func (ss *injectServerStream) SetHeader(md metadata.MD) error {
	return ss.ss.SetHeader(md)
}

func (ss *injectServerStream) SendHeader(md metadata.MD) error {
	return ss.ss.SendHeader(md)
}

func (ss *injectServerStream) SetTrailer(md metadata.MD) {
	ss.ss.SetTrailer(md)
}

func (ss *injectServerStream) Context() context.Context {
	return ss.ctx
}

func (ss *injectServerStream) SendMsg(m interface{}) error {
	return ss.ss.SendMsg(m)
}

func (ss *injectServerStream) RecvMsg(m interface{}) error {
	return ss.ss.RecvMsg(m)
}
