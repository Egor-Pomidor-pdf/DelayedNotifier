package rabbitpublisher

import (
	"fmt"

	"github.com/Egor-Pomidor-pdf/DelayedNotifier/internal/config"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

func GetRabbitProducer(rabbitCfg config.RabbitMQConfig, rabbitmqRetryStrategy retry.Strategy) (*rabbitmq.Publisher, *rabbitmq.Channel, error) {
	// init connnect
	rabbitConnect, err := rabbitmq.Connect(fmt.Sprintf("amqp://%s:%s@%s:%d/%s", rabbitCfg.User,
		rabbitCfg.Password,
		rabbitCfg.Host,
		rabbitCfg.Port,
		rabbitCfg.VHost), rabbitmqRetryStrategy.Attempts, rabbitmqRetryStrategy.Delay)

	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to rabbitmq: %w", err)
	}

	// get chanell to bind
	rabbitChannel, err := rabbitConnect.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("error connecting to rabbitmq (conn.Channel()): %w", err)
	}

	// bind channel to exchange with type direct
	rabbitExchange := rabbitmq.NewExchange(rabbitCfg.Exchange, "direct")
	err = rabbitExchange.BindToChannel(rabbitChannel)
	if err != nil {
		return nil, nil, fmt.Errorf("error binding rabbitmq channel to exchange '%s': %w",
			rabbitCfg.Exchange, err)
	}

	// declare queues
	rabbitQueneManager := rabbitmq.NewQueueManager(rabbitChannel)
	err = retry.Do(func() error {
		_, errQ := rabbitQueneManager.DeclareQueue(rabbitCfg.Queue)
		return errQ
	}, rabbitmqRetryStrategy,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("error declaring queue '%s': %w", rabbitCfg.Queue, err)
	}

	rabbitPublisher := rabbitmq.NewPublisher(rabbitChannel, rabbitCfg.Exchange)
	return rabbitPublisher, rabbitChannel, nil

}
