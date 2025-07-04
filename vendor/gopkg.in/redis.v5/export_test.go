package redis

import (
	"time"

	"gopkg.in/redis.v5/internal/pool"
)

func (c *baseClient) Pool() pool.Pooler {
	return c.connPool
}

func (c *PubSub) Pool() pool.Pooler {
	return c.base.connPool
}

func (c *PubSub) ReceiveMessageTimeout(timeout time.Duration) (*Message, error) {
	return c.receiveMessage(timeout)
}

func (c *ClusterClient) SlotAddrs(slot int) []string {
	var addrs []string
	for _, n := range c.state().slotNodes(slot) {
		addrs = append(addrs, n.Client.getAddr())
	}
	return addrs
}

// SwapSlot swaps a slot's master/slave address for testing MOVED redirects.
func (c *ClusterClient) SwapSlotNodes(slot int) []string {
	nodes := c.state().slots[slot]
	nodes[0], nodes[1] = nodes[1], nodes[0]
	return c.SlotAddrs(slot)
}
