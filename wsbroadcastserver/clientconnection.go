// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/nitro/blob/master/LICENSE

package wsbroadcastserver

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/offchainlabs/nitro/arbutil"

	"github.com/gammazero/deque"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/mailru/easygo/netpoll"
	"github.com/offchainlabs/nitro/util/stopwaiter"
)

// ClientConnection represents client connection.
type ClientConnection struct {
	stopwaiter.StopWaiter

	ioMutex  sync.Mutex
	conn     net.Conn
	creation time.Time
	clientIp net.IP

	desc            *netpoll.Desc
	Name            string
	clientManager   *ClientManager
	requestedSeqNum arbutil.MessageIndex

	lastHeardUnix int64
	out           chan []byte

	compression bool
	flateReader *wsflate.Reader

	delay          time.Duration
	delayDecayRate float64
}

func NewClientConnection(
	conn net.Conn,
	desc *netpoll.Desc,
	clientManager *ClientManager,
	requestedSeqNum arbutil.MessageIndex,
	connectingIP net.IP,
	compression bool,
	delay time.Duration,
	delayDecayRate float64,
) *ClientConnection {
	return &ClientConnection{
		conn:            conn,
		clientIp:        connectingIP,
		desc:            desc,
		creation:        time.Now(),
		Name:            fmt.Sprintf("%s@%s-%d", connectingIP, conn.RemoteAddr(), rand.Intn(10)),
		clientManager:   clientManager,
		requestedSeqNum: requestedSeqNum,
		lastHeardUnix:   time.Now().Unix(),
		out:             make(chan []byte, clientManager.config().MaxSendQueue),
		compression:     compression,
		flateReader:     NewFlateReader(),
		delay:           delay,
		delayDecayRate:  delayDecayRate,
	}
}

func (cc *ClientConnection) Age() time.Duration {
	return time.Since(cc.creation)
}

func (cc *ClientConnection) Compression() bool {
	return cc.compression
}

type msgWithScheduledTime struct {
	data  []byte
	delay time.Duration
}

type delayQueue struct {
	sync.RWMutex
	deque   deque.Deque[msgWithScheduledTime]
	pending chan struct{}
}

func NewDelayQueue() *delayQueue {
	return &delayQueue{pending: make(chan struct{})}
}

func (d *delayQueue) Front() msgWithScheduledTime {
	d.RLock()
	defer d.RUnlock()
	return d.deque.Front()
}

func (d *delayQueue) PopFront() msgWithScheduledTime {
	d.Lock()
	defer d.Unlock()
	return d.deque.PopFront()
}

func (d *delayQueue) PushBack(msg msgWithScheduledTime) {
	d.Lock()
	defer d.Unlock()
	d.deque.PushBack(msg)
	select {
	case d.pending <- struct{}{}:
	default:
	}
}

func (d *delayQueue) Len() int {
	d.RLock()
	defer d.RUnlock()
	return d.deque.Len()
}

func (cc *ClientConnection) Start(parentCtx context.Context) {
	cc.StopWaiter.Start(parentCtx, cc)

	delayQueue := NewDelayQueue()

	cc.LaunchThread(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case data := <-cc.out:
				delayQueue.PushBack(
					msgWithScheduledTime{
						data,
						cc.delay,
					})

				if cc.delay != 0 {
					cc.delay = time.Duration(float64(cc.delay) * cc.delayDecayRate)
					if cc.delay == 0 {
						log.Trace("Client now connected without delay", "client", cc.Name)
					}
				}
			}
		}
	})

	cc.LaunchThread(func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-delayQueue.pending:
				for delayQueue.Len() != 0 {
					msg := delayQueue.PopFront()

					time.Sleep(msg.delay)

					err := cc.writeRaw(msg.data)
					if err != nil {
						logWarn(err, fmt.Sprintf("Error writing data to client %s", cc.Name))
						cc.clientManager.Remove(cc)
					}
				}
			}
		}
	})
}

func (cc *ClientConnection) StopOnly() { // Ignore errors from conn.Close since we are just shutting down
	_ = cc.conn.Close()
	if cc.Started() {
		cc.StopWaiter.StopOnly()
	}
}

func (cc *ClientConnection) RequestedSeqNum() arbutil.MessageIndex {
	return cc.requestedSeqNum
}

func (cc *ClientConnection) GetLastHeard() time.Time {
	return time.Unix(atomic.LoadInt64(&cc.lastHeardUnix), 0)
}

// Receive reads next message from client's underlying connection.
// It blocks until full message received.
func (cc *ClientConnection) Receive(ctx context.Context, timeout time.Duration) ([]byte, ws.OpCode, error) {
	msg, op, err := cc.readRequest(ctx, timeout)
	if err != nil {
		_ = cc.conn.Close()
		return nil, op, err
	}

	return msg, op, err
}

// readRequests reads json-rpc request from connection.
func (cc *ClientConnection) readRequest(ctx context.Context, timeout time.Duration) ([]byte, ws.OpCode, error) {
	cc.ioMutex.Lock()
	defer cc.ioMutex.Unlock()

	atomic.StoreInt64(&cc.lastHeardUnix, time.Now().Unix())

	var data []byte
	var opCode ws.OpCode
	var err error
	data, opCode, err = ReadData(ctx, cc.conn, nil, timeout, ws.StateServerSide, cc.compression, cc.flateReader)
	return data, opCode, err
}

func (cc *ClientConnection) Write(x interface{}) error {
	cc.ioMutex.Lock()
	defer cc.ioMutex.Unlock()

	notCompressed, compressed, err := serializeMessage(cc.clientManager, x, !cc.compression, cc.compression)
	if err != nil {
		return err
	}

	if cc.compression {
		cc.out <- compressed.Bytes()
	} else {
		cc.out <- notCompressed.Bytes()
	}
	return nil
}

func (cc *ClientConnection) writeRaw(p []byte) error {
	cc.ioMutex.Lock()
	defer cc.ioMutex.Unlock()

	_, err := cc.conn.Write(p)

	return err
}

func (cc *ClientConnection) Ping() error {
	cc.ioMutex.Lock()
	defer cc.ioMutex.Unlock()
	_, err := cc.conn.Write(ws.CompiledPing)
	if err != nil {
		return err
	}

	return nil
}
