package mail

import (
	"fmt"
	"time"

	context "golang.org/x/net/context"
)

const en_fmt = "2006-01-02 15:04:05.999999"

var Mail_serviceDesc = _Mail_serviceDesc

type Email struct {
	HostName   string
	ListenPort int
}

func (this *Email) to_string() string {
	return fmt.Sprintf("%s-%d", this.HostName, this.ListenPort)
}

func (this *Email) Call(
	ctx context.Context, sender *Sender) (*Reciver, error) {
	msg := time.Now().Format(en_fmt)
	return &Reciver{Reply: fmt.Sprintf(
		"From Email server %s: Hello, now is %s", this.to_string(), msg)}, nil
}
