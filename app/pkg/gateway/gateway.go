package gateway

import (
	"fmt"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	"github.com/hyperledger/fabric-gateway/pkg/identity"
	"google.golang.org/grpc"
	"time"
)

var (
	EvaluateTimeout     = 5 * time.Second
	EndorsementTimeout  = 15 * time.Second
	SubmitTimeout       = 5 * time.Second
	CommitStatusTimeout = 1 * time.Minute
)

func NewGateway(id identity.Identity, sign identity.Sign, connection *grpc.ClientConn) (*client.Gateway, error) {
	gateway, err := client.Connect(
		id,
		client.WithSign(sign),
		client.WithClientConnection(connection),
		client.WithEvaluateTimeout(EvaluateTimeout),
		client.WithEndorseTimeout(EndorsementTimeout),
		client.WithSubmitTimeout(SubmitTimeout),
		client.WithCommitStatusTimeout(CommitStatusTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create a gateway: %w", err)
	}

	return gateway, nil
}
