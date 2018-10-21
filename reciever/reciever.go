package reciever

import (
    "net"
    "context"
    "github.com/pkg/errors"
    "google.golang.org/grpc"
    "google.golang.org/grpc/peer"
    "github.com/potix/log_monitor/configurator"
    logpb "github.com/potix/log_monitor/logpb"
)

// Reciever is reciever
type Reciever struct {
    listen net.Listener
    server *grpc.Server
    config *configurator.LogRecieverConfig
}

func (r *Reciever) getRemoteAddr(ctx context.Context) (string) {
    peer, ok := peer.FromContext(ctx)
    if !ok {
        return "NoPeer"
    }
    tcpAddr, ok := peer.Addr.(*net.TCPAddr);
    if !ok {
        return "NotTCP"
    }
    return tcpAddr.IP.String()
}

// Transfer is transfer
func (r *Reciever) Transfer(ctx context.Context, request *logpb.TransferRequest) (*logpb.TransferReply, error) {
     addr := r.getRemoteAddr(ctx)
     err := logstore.Save(ctx, addr, request)
     if err != nil {
        return &logpb.TransferReply{
            Success: false,
            Msg: err.Error(),
        }, errors.Wrapf(err, "can not save log (%v, %v, %v, %v)", request.Label, request.Host, addr, request.Path)
     }
     return &logpb.TransferReply{
          Success: true,
          Msg: "OK",
     }, nil
}

// Start is start
func (r *Reciever) Start() (error) {
    err := r.server.Serve(r.listen)
    if err != nil { 
        return errors.Wrap(err, "can not start server")
    }
    return nil
}

// Stop is stop
func (r *Reciever) Stop() {
     r.server.GracefulStop()
}

// NewReciever is create new reciver
func NewReciever(config *configurator.LogRecieverConfig) (error){
    listen, err := net.Listen("tcp", config.AddrPort)
    if err != nil {
        return errors.Wrapf(err, "can not listen addr port (%v)", config.AddrPort)
    }
    server := grpc.NewServer()
    reciever := &Reciever{
        listen: listen,
        server: server,
        config: config,
    }
    logpb.RegisterLogServer(server, reciever)

    return nil
}
