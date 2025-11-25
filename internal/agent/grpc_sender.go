package agent

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/GoLessons/go-musthave-metrics/internal/model"
	"github.com/GoLessons/go-musthave-metrics/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type grpcSender struct {
	conn    *grpc.ClientConn
	client  proto.MetricsClient
	address string
	realIP  string
	timeout time.Duration
}

func NewGRPCSender(address string) Sender {
	conn, _ := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := proto.NewMetricsClient(conn)

	ip := ""
	if c, err := net.Dial("udp", address); err == nil {
		if la, ok := c.LocalAddr().(*net.UDPAddr); ok && la.IP != nil {
			ip = la.IP.String()
		}
		_ = c.Close()
	}
	if ip == "" {
		ip = "127.0.0.1"
	}

	return &grpcSender{
		conn:    conn,
		client:  client,
		address: address,
		realIP:  ip,
		timeout: 5 * time.Second,
	}
}

func (s *grpcSender) Send(metric model.Metrics) error {
	pm, err := s.modelToProto(metric)
	if err != nil {
		return err
	}
	return s.sendUpdate([]*proto.Metric{pm})
}

func (s *grpcSender) SendBatch(metrics []model.Metrics) error {
	list := make([]*proto.Metric, 0, len(metrics))
	for _, m := range metrics {
		pm, err := s.modelToProto(m)
		if err != nil {
			return err
		}
		list = append(list, pm)
	}
	return s.sendUpdate(list)
}

func (s *grpcSender) Close() {
	_ = s.conn.Close()
}

func (s *grpcSender) sendUpdate(list []*proto.Metric) error {
	md := metadata.Pairs("x-real-ip", s.realIP)
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	ctx = metadata.NewOutgoingContext(ctx, md)
	_, err := s.client.UpdateMetrics(ctx, &proto.UpdateMetricsRequest{Metrics: list})
	return err
}

func (s *grpcSender) modelToProto(m model.Metrics) (*proto.Metric, error) {
	pm := &proto.Metric{Id: m.ID}
	switch m.MType {
	case model.Gauge:
		if m.Value == nil {
			return nil, fmt.Errorf("gauge metric %s has nil Value", m.ID)
		}
		pm.Type = proto.Metric_GAUGE
		pm.Value = *m.Value
	case model.Counter:
		if m.Delta == nil {
			return nil, fmt.Errorf("counter metric %s has nil Delta", m.ID)
		}
		pm.Type = proto.Metric_COUNTER
		pm.Delta = *m.Delta
	default:
		return nil, fmt.Errorf("unsupported metric type: %s", m.MType)
	}
	return pm, nil
}
