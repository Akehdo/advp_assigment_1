package messaging

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func ConnectNATS(url, serviceName string, logger *log.Logger) (*nats.Conn, error) {
	if logger == nil {
		logger = log.Default()
	}

	conn, err := nats.Connect(
		url,
		nats.Name(serviceName),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			logger.Printf("%s disconnected from NATS: %v", serviceName, err)
		}),
		nats.ReconnectHandler(func(conn *nats.Conn) {
			logger.Printf("%s reconnected to NATS at %s", serviceName, conn.ConnectedUrl())
		}),
		nats.ClosedHandler(func(conn *nats.Conn) {
			if err := conn.LastError(); err != nil {
				logger.Printf("%s closed NATS connection: %v", serviceName, err)
				return
			}

			logger.Printf("%s closed NATS connection", serviceName)
		}),
		nats.ErrorHandler(func(_ *nats.Conn, sub *nats.Subscription, err error) {
			if sub != nil {
				logger.Printf("%s NATS subscription error on %s: %v", serviceName, sub.Subject, err)
				return
			}

			logger.Printf("%s NATS error: %v", serviceName, err)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to NATS: %w", err)
	}

	return conn, nil
}
