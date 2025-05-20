package discovery

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	mdns "github.com/hashicorp/mdns"
)

// Service advertises and discovers peers via mDNS.
type Service struct {
	ServiceName string
	Port        int
}

func NewService(name string, port int) *Service {
	return &Service{
		ServiceName: name,
		Port:        port,
	}
}

func (s *Service) Start(ctx context.Context) error {
	host, _ := os.Hostname()
	info := []string{"nebula-edge"}
	service, err := mdns.NewMDNSService(host, s.ServiceName, "", "", s.Port, nil, info)
	if err != nil {
		return err
	}
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		server.Shutdown()
	}()
	log.Printf("mDNS service started (%s)", s.ServiceName)
	return nil
}

// DiscoverPeers returns addresses discovered via mDNS.
func (s *Service) DiscoverPeers(ctx context.Context) ([]string, error) {
	var entriesCh = make(chan *mdns.ServiceEntry, 4)
	var peers []string
	go func() {
		for entry := range entriesCh {
			addr := net.JoinHostPort(entry.AddrV4.String(), fmt.Sprintf("%d", entry.Port))
			peers = append(peers, addr)
		}
	}()
	mdns.Lookup(s.ServiceName, entriesCh)
	select {
	case <-time.After(2 * time.Second):
	case <-ctx.Done():
	}
	close(entriesCh)
	return peers, nil
}
