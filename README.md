# etcd-mini-cluster

This project provides example code to create an embedded [etcd](https://etcd.io) cluster (MiniCluster)
in your own Golang program.

People could use such an etcd cluster as a key value storage without depending on any extra database.

This MiniCluster could also be used in unit test code to create an etcd cluster.

# How to Use

Define a set of ports to use, then call `StartMiniCluster`:

```
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
```


See [this test](https://github.com/datapunchorg/etcd-min-cluster/blob/main/pkg/etcdex/minicluster_test.go) for detailed code example.
