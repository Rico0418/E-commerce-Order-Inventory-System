package config

import (
	"log"
	"os"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitConn  *amqp.Connection
	rabbitOnce  sync.Once
	rabbitMutex sync.Mutex
)

func GetRabbitConn() *amqp.Connection {
	rabbitOnce.Do(func() {
		url := os.Getenv("RABBITMQ_URL")
		if url == "" {
			log.Fatal("RABBITMQ_URL is required")
		}
		conn, err := amqp.Dial(url)
		if err != nil {
			log.Fatalf("failed to connect to rabbitmq: %v", err)
		}
		rabbitConn = conn
		log.Println("connected to rabbitmq")
	})
	return rabbitConn
}

func NewChannel() (*amqp.Channel, error) {
	conn := GetRabbitConn()
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func EnsureDirectExchange(ch *amqp.Channel, exchangeName string) error {
	return ch.ExchangeDeclare(
		exchangeName, 
		"direct",     
		true,         
		false,        
		false,        
		false,        
		nil,          
	)
}


func DeclareQuorumQueue(ch *amqp.Channel, queueName, exchangeName, routingKey string) (string, error) {
	args := amqp.Table{
		"x-queue-type": "quorum",
	}
	q, err := ch.QueueDeclare(
		queueName,
		true,  
		false, 
		false, 
		false, 
		args,  
	)
	if err != nil {
		return "", err
	}
	if err := ch.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil {
		return "", err
	}
	return q.Name, nil
}


func PublishJSON(ch *amqp.Channel, exchange, routingKey string, body []byte) error {
	rabbitMutex.Lock()
	defer rabbitMutex.Unlock()
	return ch.Publish(
		exchange,
		routingKey,
		false, 
		false, 
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
