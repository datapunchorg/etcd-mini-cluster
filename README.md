# etcd-mini-cluster

This project provides example code to create an embedded [etcd](https://etcd.io) cluster (MiniCluster)
in your own Golang program.

People could use such an etcd cluster as a key value storage without depending on any extra database.

This MiniCluster could also be used in unit test code to create an etcd cluster.

# How to Use

Define a set of ports to use, then call `StartMiniCluster`:

```
import (
	"github.com/datapunchorg/etcd-mini-cluster/pkg/etcdex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMiniCluster(t *testing.T) {
	ports := []etcdex.EtcdListenPortPair{
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
	miniCluster, err := etcdex.StartMiniCluster(ports)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	defer miniCluster.Stop()
}
```


See [this test](https://github.com/datapunchorg/etcd-min-cluster/blob/main/pkg/etcdex/minicluster_test.go) for detailed code example.
