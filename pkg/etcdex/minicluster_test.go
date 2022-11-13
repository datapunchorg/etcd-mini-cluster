package etcdex

import (
	"context"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func TestMiniCluster(t *testing.T) {
	ports := []EtcdListenPortPair{
		{
			ListenPeerPort:   2380,
			ListenClientPort: 2379,
		},
		{
			ListenPeerPort:   2382,
			ListenClientPort: 2381,
		},
		{
			ListenPeerPort:   2384,
			ListenClientPort: 2383,
		},
	}
	miniCluster, err := StartMiniCluster(ports)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	defer miniCluster.Stop()

	dialTimeout := 10 * time.Second
	requestTimeout := 30 * time.Second
	endpoints := miniCluster.GetClientEndpoints()

	ctx, _ := context.WithTimeout(context.Background(), requestTimeout)
	client, err := clientv3.New(clientv3.Config{
		DialTimeout: dialTimeout,
		Endpoints:   endpoints,
	})
	assert.Nil(t, err)

	kv := clientv3.NewKV(client)
	key := "hello"
	_, err = kv.Put(ctx, key, "world")
	assert.Nil(t, err)

	getResponse, err := kv.Get(ctx, key)
	assert.Nil(t, err)
	assert.Equal(t, "world", string(getResponse.Kvs[0].Value))

	client.Close()
}
