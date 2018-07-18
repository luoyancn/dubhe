package msg

import (
	"fmt"
	"time"

	context "golang.org/x/net/context"
)

const en_fmt = "2006-01-02 15:04:05.999999"

var Messages_serviceDesc = _Messages_serviceDesc

type Service struct {
	HostName   string
	ListenPort int
}

func (this *Service) to_string() string {
	return fmt.Sprintf("%s-%d", this.HostName, this.ListenPort)
}

func (this *Service) Call(
	ctx context.Context, req *Request) (*Response, error) {
	msg := time.Now().Format(en_fmt)
	return &Response{Resp: fmt.Sprintf(
		"From grpc server %s: Hello, now is %s", this.to_string(), msg)}, nil
}
