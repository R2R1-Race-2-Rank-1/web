package redis_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gopkg.in/redis.v5"
)

var _ = Describe("pool", func() {
	var client *redis.Client

	BeforeEach(func() {
		client = redis.NewClient(redisOptions())
		Expect(client.FlushDb().Err()).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(client.Close()).NotTo(HaveOccurred())
	})

	It("respects max size", func() {
		perform(1000, func(id int) {
			val, err := client.Ping().Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(Equal("PONG"))
		})

		pool := client.Pool()
		Expect(pool.Len()).To(BeNumerically("<=", 10))
		Expect(pool.FreeLen()).To(BeNumerically("<=", 10))
		Expect(pool.Len()).To(Equal(pool.FreeLen()))
	})

	It("respects max size on multi", func() {
		perform(1000, func(id int) {
			var ping *redis.StatusCmd

			err := client.Watch(func(tx *redis.Tx) error {
				cmds, err := tx.Pipelined(func(pipe *redis.Pipeline) error {
					ping = pipe.Ping()
					return nil
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(cmds).To(HaveLen(1))
				return err
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(ping.Err()).NotTo(HaveOccurred())
			Expect(ping.Val()).To(Equal("PONG"))
		})

		pool := client.Pool()
		Expect(pool.Len()).To(BeNumerically("<=", 10))
		Expect(pool.FreeLen()).To(BeNumerically("<=", 10))
		Expect(pool.Len()).To(Equal(pool.FreeLen()))
	})

	It("respects max size on pipelines", func() {
		perform(1000, func(id int) {
			pipe := client.Pipeline()
			ping := pipe.Ping()
			cmds, err := pipe.Exec()
			Expect(err).NotTo(HaveOccurred())
			Expect(cmds).To(HaveLen(1))
			Expect(ping.Err()).NotTo(HaveOccurred())
			Expect(ping.Val()).To(Equal("PONG"))
			Expect(pipe.Close()).NotTo(HaveOccurred())
		})

		pool := client.Pool()
		Expect(pool.Len()).To(BeNumerically("<=", 10))
		Expect(pool.FreeLen()).To(BeNumerically("<=", 10))
		Expect(pool.Len()).To(Equal(pool.FreeLen()))
	})

	It("respects max size on pubsub", func() {
		connPool := client.Pool()

		perform(1000, func(id int) {
			pubsub, err := client.Subscribe()
			Expect(err).NotTo(HaveOccurred())
			Expect(pubsub.Close()).NotTo(HaveOccurred())
		})

		Expect(connPool.Len()).To(Equal(connPool.FreeLen()))
		Expect(connPool.Len()).To(BeNumerically("<=", 10))
	})

	It("removes broken connections", func() {
		cn, _, err := client.Pool().Get()
		Expect(err).NotTo(HaveOccurred())
		cn.SetNetConn(&badConn{})
		Expect(client.Pool().Put(cn)).NotTo(HaveOccurred())

		err = client.Ping().Err()
		Expect(err).To(MatchError("bad connection"))

		val, err := client.Ping().Result()
		Expect(err).NotTo(HaveOccurred())
		Expect(val).To(Equal("PONG"))

		pool := client.Pool()
		Expect(pool.Len()).To(Equal(1))
		Expect(pool.FreeLen()).To(Equal(1))

		stats := pool.Stats()
		Expect(stats.Requests).To(Equal(uint32(4)))
		Expect(stats.Hits).To(Equal(uint32(2)))
		Expect(stats.Timeouts).To(Equal(uint32(0)))
	})

	It("reuses connections", func() {
		for i := 0; i < 100; i++ {
			val, err := client.Ping().Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(Equal("PONG"))
		}

		pool := client.Pool()
		Expect(pool.Len()).To(Equal(1))
		Expect(pool.FreeLen()).To(Equal(1))

		stats := pool.Stats()
		Expect(stats.Requests).To(Equal(uint32(101)))
		Expect(stats.Hits).To(Equal(uint32(100)))
		Expect(stats.Timeouts).To(Equal(uint32(0)))
	})

	It("removes idle connections", func() {
		stats := client.PoolStats()
		Expect(stats).To(Equal(&redis.PoolStats{
			Requests:   1,
			Hits:       0,
			Timeouts:   0,
			TotalConns: 1,
			FreeConns:  1,
		}))

		time.Sleep(2 * time.Second)

		stats = client.PoolStats()
		Expect(stats).To(Equal(&redis.PoolStats{
			Requests:   1,
			Hits:       0,
			Timeouts:   0,
			TotalConns: 0,
			FreeConns:  0,
		}))
	})
})
