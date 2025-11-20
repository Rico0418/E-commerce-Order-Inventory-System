package notification

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"ecommerce-app/events"
	"ecommerce-app/config"

)

func StartNotificationWorker(ctx context.Context) error {
	ch, err := config.NewChannel()
	if err != nil {
		return err
	}

	exchange := os.Getenv("RABBITMQ_EXCHANGE")
	if exchange == "" {
		exchange = "orders_direct"
	}
	confirmRK := os.Getenv("RABBITMQ_CONFIRM_ROUTING_KEY")
	if confirmRK == "" {
		confirmRK = "order.confirmed"
	}
	failedRK := os.Getenv("RABBITMQ_FAILED_ROUTING_KEY")
	if failedRK == "" {
		failedRK = "order.failed"
	}
	confirmQueue := os.Getenv("RABBITMQ_CONFIRM_QUEUE")
	if confirmQueue == "" {
		confirmQueue = "order_confirmed_queue"
	}
	failedQueue := os.Getenv("RABBITMQ_FAILED_QUEUE")
	if failedQueue == "" {
		failedQueue = "order_failed_queue"
	}

	if err := config.EnsureDirectExchange(ch, exchange); err != nil {
		return err
	}
	_, _ = config.DeclareQuorumQueue(ch, confirmQueue, exchange, confirmRK)
	_, _ = config.DeclareQuorumQueue(ch, failedQueue, exchange, failedRK)

	confirmMsgs, err := ch.Consume(confirmQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	failedMsgs, err := ch.Consume(failedQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	log.Println("notification worker: consuming confirmed & failed queues")

	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = ch.Close()
				return
			case d, ok := <-confirmMsgs:
				if !ok {
					return
				}
				var payload events.OrderResultPayload
				if err := json.Unmarshal(d.Body, &payload); err != nil {
					log.Printf("notification: invalid confirmed payload: %v", err)
					d.Nack(false, false)
					continue
				}
				log.Printf("notification: sending CONFIRM email for order %s to user %s", payload.OrderID, payload.UserID)
				time.Sleep(200 * time.Millisecond) 
				log.Printf("notification: CONFIRM email sent for order %s", payload.OrderID)
				d.Ack(false)
			}
		}
	}()

	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = ch.Close()
				return
			case d, ok := <-failedMsgs:
				if !ok {
					return
				}
				var payload events.OrderResultPayload
				if err := json.Unmarshal(d.Body, &payload); err != nil {
					log.Printf("notification: invalid failed payload: %v", err)
					d.Nack(false, false)
					continue
				}
				log.Printf("notification: sending CANCEL email for order %s to user %s (reason=%s)", payload.OrderID, payload.UserID, payload.Reason)
				time.Sleep(200 * time.Millisecond)
				log.Printf("notification: CANCEL email sent for order %s", payload.OrderID)
				d.Ack(false)
			}
		}
	}()

	return nil
}
