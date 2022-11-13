package etcdex

import (
	"fmt"
	"go.etcd.io/etcd/server/v3/embed"
	"log"
	"net/url"
	"os"
	"time"
)

type ServerConfig struct {
	ListenPeerPort   int // e.g. 2380
	ListenClientPort int // e.g. 2379
	LocalServerName  string
	InitialCluster   string
	RootDir          string
}

type MiniServer struct {
	ServerId  int
	Config    ServerConfig
	etcd      *embed.Etcd
	stoppedCh chan any
}

func StartMiniServer(serverId int, config ServerConfig) (*MiniServer, error) {
	if config.LocalServerName == "" {
		config.LocalServerName = "localhost"
	}

	if config.RootDir == "" {
		rootDir, err := os.MkdirTemp("", "test")
		if err != nil {
			return nil, fmt.Errorf("failed to make temp dir: %s", err.Error())
		}
		log.Printf("Using root dir %s for etcd server", rootDir)
		config.RootDir = rootDir
	}

	s := MiniServer{
		ServerId:  serverId,
		Config:    config,
		stoppedCh: make(chan any),
	}

	err := s.Start()
	if err != nil {
		return nil, err
	}

	// TODO find other way to wait for server ready without sleeping
	time.Sleep(500 * time.Millisecond)

	return &s, nil
}

func (s *MiniServer) GetClientEndpoint() string {
	return fmt.Sprintf("%s:%d", s.Config.LocalServerName, s.Config.ListenClientPort)
}

func (s *MiniServer) Start() error {
	etcdName := fmt.Sprintf("%d", s.ServerId)

	etcdConfig := embed.NewConfig()
	etcdConfig.Name = etcdName
	etcdConfig.Dir = fmt.Sprintf("%s/%s", s.Config.RootDir, etcdName)
	etcdConfig.LPUrls = parseUrls([]string{fmt.Sprintf("http://0.0.0.0:%d", s.Config.ListenPeerPort)})                        // listen-peer-urls
	etcdConfig.LCUrls = parseUrls([]string{fmt.Sprintf("http://0.0.0.0:%d", s.Config.ListenClientPort)})                      // listen-client-urls
	etcdConfig.APUrls = parseUrls([]string{fmt.Sprintf("http://%s:%d", s.Config.LocalServerName, s.Config.ListenPeerPort)})   // advertise-peer-urls
	etcdConfig.ACUrls = parseUrls([]string{fmt.Sprintf("http://%s:%d", s.Config.LocalServerName, s.Config.ListenClientPort)}) // advertise-client-urls
	etcdConfig.InitialCluster = s.Config.InitialCluster                                                                       // e.g. 0=http://localhost:2380

	etcd, err := embed.StartEtcd(etcdConfig)
	if err != nil {
		s.Stop()
		return fmt.Errorf("failed to start etcd: %s", err.Error())
	}

	select {
	case <-etcd.Server.ReadyNotify():
		log.Printf("Etcd server is ready!")
	case <-time.After(60 * time.Second):
		etcd.Server.Stop()
		etcd.Close()
		return fmt.Errorf("etcd server timed out and stopped")
	}

	s.etcd = etcd

	go func() {
		select {
		case err = <-etcd.Err():
			if err != nil {
				log.Printf("[WARN] etcd error: %s", err.Error())
			} else {
				log.Printf("Exit etcd waiting go routine")
			}
		case <-s.stoppedCh:
			log.Printf("Exit etcd waiting go routine due to being stopped")
		}
	}()

	log.Printf("Broker server started")
	return nil
}

func (s *MiniServer) Stop() {
	log.Printf("Stopping etcd server")

	if s.etcd != nil {
		log.Printf("Stopping etcd server")
		s.etcd.Server.Stop()
		s.etcd.Close()
	}

	close(s.stoppedCh)
}

func parseUrls(values []string) []url.URL {
	urls := make([]url.URL, 0, len(values))
	for _, s := range values {
		u, err := url.Parse(s)
		if err != nil {
			log.Printf("Invalid url %s: %s", s, err.Error())
			continue
		}
		urls = append(urls, *u)
	}
	return urls
}
