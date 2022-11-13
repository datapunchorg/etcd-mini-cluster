package etcdex

import (
	"fmt"
	"log"
	"strings"
)

type MiniCluster struct {
	Ports               []EtcdListenPortPair
	localServerName     string
	servers             []*MiniServer
	serverStartResultCh chan startServerResult
}

type EtcdListenPortPair struct {
	ListenPeerPort   int
	ListenClientPort int
}

type startServerResult struct {
	serverId int
	server   *MiniServer
	err      error
}

func StartMiniCluster(ports []EtcdListenPortPair) (*MiniCluster, error) {
	m := MiniCluster{
		Ports:               ports,
		serverStartResultCh: make(chan startServerResult, len(ports)),
	}

	if len(m.Ports) == 0 {
		m.Ports = []EtcdListenPortPair{
			{
				ListenPeerPort:   2380,
				ListenClientPort: 2379,
			},
		}
	}
	if m.localServerName == "" {
		m.localServerName = "localhost"
	}

	err := m.Start()
	if err != nil {
		m.Stop()
		return nil, err
	}
	return &m, nil
}

func (m *MiniCluster) Start() error {
	localHostName := m.localServerName

	initialClusterItems := make([]string, len(m.Ports))
	for i, p := range m.Ports {
		serverId := i
		initialClusterItems[i] = fmt.Sprintf("%d=http://%s:%d", serverId, localHostName, p.ListenPeerPort)
	}
	initialCluster := strings.Join(initialClusterItems, ",")

	m.servers = make([]*MiniServer, len(m.Ports))
	for i, p := range m.Ports {
		serverId := i
		config := ServerConfig{
			ListenPeerPort:   p.ListenPeerPort,
			ListenClientPort: p.ListenClientPort,
			InitialCluster:   initialCluster,
		}
		go func() {
			server, err := StartMiniServer(serverId, config)
			m.serverStartResultCh <- startServerResult{
				serverId: serverId,
				server:   server,
				err:      err,
			}
		}()
	}

	var lastErr error
	for _ = range m.Ports {
		serverStartResult := <-m.serverStartResultCh
		m.servers[serverStartResult.serverId] = serverStartResult.server
		if serverStartResult.err != nil {
			lastErr = serverStartResult.err
			log.Printf("[WARN] Failed to start ectd server %d: %s", serverStartResult.serverId, serverStartResult.err.Error())
		}
	}

	if lastErr != nil {
		m.Stop()
		return lastErr
	}

	return nil
}

func (m *MiniCluster) Stop() {
	for _, server := range m.servers {
		if server != nil {
			server.Stop()
		}
	}
}

func (m *MiniCluster) GetClientEndpoints() []string {
	result := make([]string, len(m.Ports))
	for i, p := range m.Ports {
		result[i] = fmt.Sprintf("%s:%d", m.localServerName, p.ListenClientPort)
	}
	return result
}
