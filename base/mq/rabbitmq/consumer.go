package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type ConsumerConfig struct {
	// 交换机名
	ExchangeName string
	// 交换机类型
	ExchangeType string
	// 队列绑定交换机
	BindingKey string
	// 交换机名
	QueueName string
	// 消息代理重启后，队列是否依旧存在
	Durable bool
	// 当最后一个消费者退订后即被删除
	AutoDelete bool
	Internal   bool
	// 只被一个连接（connection）使用，而且当连接关闭后队列即被删除
	Exclusive bool
	// 是否等待服务器确认请求并立即开始传送
	// 如果无法消费，则需要一个渠道将引发异常并关闭通道。
	NoWait bool
	// 是否自动ACK
	AutoAck bool
	Tag     string
}

type Consumer struct {
	config *ConsumerConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
	queue  amqp.Queue
}

func NewConsumer(uri string, config *ConsumerConfig) (*Consumer, error) {
	c := &Consumer{
		config: config,
	}

	var err error

	c.conn, err = amqp.Dial(uri)
	if err != nil {
		return nil, err
	}

	c.ch, err = c.conn.Channel()
	if err != nil {
		return nil, err
	}

	err = c.ch.ExchangeDeclare(
		c.config.ExchangeName,
		c.config.ExchangeType,
		c.config.Durable,
		c.config.AutoDelete,
		c.config.Internal,
		c.config.NoWait,
		nil,
	)
	if err != nil {
		return nil, err
	}

	c.queue, err = c.ch.QueueDeclare(
		c.config.QueueName,
		c.config.Durable,
		c.config.AutoDelete,
		c.config.Exclusive,
		c.config.NoWait,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Consumer) Consumer() (<-chan amqp.Delivery, error) {
	return c.ch.Consume(
		c.queue.Name,
		c.config.Tag,
		c.config.AutoAck,
		c.config.Exclusive,
		false,
		c.config.NoWait,
		nil,
	)
}
