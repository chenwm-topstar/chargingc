package servicecontext

import (
	"context"

	"cchome-admin-topstar/utils/lg"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// ServiceContextMetadataKey is key to used in GRPC Context.
	ServiceContextMetadataKey = "consul-tag"
)

func appendGRPCContext(ctx context.Context, sc ServiceContext) context.Context {
	if len(sc) == 0 {
		return ctx
	}
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	md.Set(ServiceContextMetadataKey, sc.Specs()...)
	return metadata.NewOutgoingContext(ctx, md)
}

func fromGRPCContext(ctx context.Context) ServiceContext {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	if len(md.Get(ServiceContextMetadataKey)) == 0 {
		return nil
	}
	return New(md.Get(ServiceContextMetadataKey))
}

// UnaryClientInterceptor injects log context into outgoing grpc request.
func UnaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = appendGRPCContext(ctx, FromContext(ctx))
	return invoker(ctx, method, req, reply, cc, opts...)
}

// StreamClientInterceptor injects log context into outgoing grpc request.
func StreamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = appendGRPCContext(ctx, FromContext(ctx))
	return streamer(ctx, desc, cc, method, opts...)
}

// UnaryServerInterceptor retrieve log context from incoming grpc request.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	sc := fromGRPCContext(ctx)
	if sc != nil {
		ctx = With(ctx, sc)
		lg.Debug("Service context", sc.Specs())
	}
	return handler(ctx, req)
}

// StreamServerInterceptor retrieve log context from incoming grpc request.
func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	sc := fromGRPCContext(ss.Context())
	if sc == nil {
		return handler(srv, ss)
	}
	// Wrap a server stream implementation to modify the context to include the data.
	ctx := With(ss.Context(), sc)
	lg.Debug("Service context", sc.Specs())
	return handler(srv, newInjectServerStream(ctx, ss))
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
