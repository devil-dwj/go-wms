package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

type ProducerConfig struct {
	// 交换机名
	ExchangeName string
	// 交换机类型
	ExchangeType string
	// 路由键
	RoutingKey string
	// 持久化(交换机在消息代理（broker）重启后是否依旧存在)
	Durable bool
	// 当所有与之绑定的消息队列都完成了对此交换机的使用后，删掉它
	AutoDelete bool
	// 内部的声明不接受外部连接
	Internal bool
	// 当noWait为true时，在不等待服务器确认的情况下进行声明。
	// 该通道可能因错误而关闭。添加NotifyClose侦听器对任何例外情况做出回应。
	NoWait bool
	// 投递模式（持久化 或 非持久化）
	Delivery uint8
}

type Producer struct {
	config *ProducerConfig
	conn   *amqp.Connection
	ch     *amqp.Channel
}

func NewProducer(uri string, config *ProducerConfig) (*Producer, error) {
	p := &Producer{}
	p.config = config

	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, err
	}
	p.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	p.ch = ch

	err = p.ch.ExchangeDeclare(
		config.ExchangeName,
		config.ExchangeType,
		config.Durable,
		config.AutoDelete,
		config.Internal,
		config.NoWait,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Producer) Publish(body []byte) error {
	err := p.ch.Publish(
		p.config.ExchangeName,
		p.config.RoutingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            body,
			DeliveryMode:    p.config.Delivery,
			Priority:        0,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (p *Producer) Close() error {
	return p.conn.Close()
}
