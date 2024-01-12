package client

import (
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type gRpcClientPool struct {
	mu          sync.RWMutex
	pool        []*connectionpool
	addr        string
	minpoolsize int
	maxpoolsize int
	poolSize    *atomic.Int32
	maxsessions int32
}

type connectionpool struct {
	connection *grpc.ClientConn
	lastCall   time.Time
	status     bool
	sessions   *atomic.Int32
}

func (proxy *gRpcClientPool) getPool() *connectionpool {
	maxRetries := 50
	retryInterval := time.Millisecond
	for i := 0; i < maxRetries; i++ {
		getconn := proxy.getConnection()
		if getconn == nil {
			if proxy.poolSize.Load() <= int32(proxy.maxpoolsize) {
				newconn := proxy.addConnection()
				if newconn == nil {
					return nil
				}
				return newconn
			}
			time.Sleep(retryInterval)
		} else if getconn != nil {
			return getconn
		}
	}
	return nil
}

func (proxy *gRpcClientPool) getConnection() *connectionpool {
	proxy.mu.RLock()
	defer proxy.mu.RUnlock()
	for i, wrapper := range proxy.pool {
		if wrapper.sessions.Load() <= proxy.maxsessions && wrapper.status {
			proxy.pool[i].lastCall = time.Now()
			return proxy.pool[i]
		}
	}
	return nil
}

func (proxy *gRpcClientPool) establishedConnection() *grpc.ClientConn {
	conn, err := grpc.Dial(proxy.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}
	return conn
}

func (proxy *gRpcClientPool) addConnection() *connectionpool {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	connection := &connectionpool{}
	conn, err := grpc.Dial(proxy.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil
	}
	connection.connection = conn
	connection.status = true
	connection.lastCall = time.Now()
	proxy.pool = append(proxy.pool, connection)
	proxy.poolSize.Add(1)
	return connection
}

func (proxy *gRpcClientPool) rollbackConnection(c *connectionpool) {
	c.sessions.Add(-1)
}

func (proxy *gRpcClientPool) removeConnection(cp *connectionpool) {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	for index, wrapper := range proxy.pool {
		if wrapper == cp {
			if index > proxy.minpoolsize {
				proxy.pool[index].connection.Close()
				proxy.pool = append(proxy.pool[:index], proxy.pool[index+1:]...)
				proxy.poolSize.Add(-1)
			} else {
				proxy.pool[index].connection.Close()
				proxy.pool[index] = proxy.addConnection()
				// already add connection, channel size+1 but this logic purpose is maintain channel size
				// because only channel connection change healthy
				proxy.poolSize.Add(-1)
			}
		}
	}
}
