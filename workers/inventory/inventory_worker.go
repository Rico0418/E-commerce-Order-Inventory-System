package inventory

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"ecommerce-app/events"
	prodEntities "ecommerce-app/domain/products/entities"
	orderEntities "ecommerce-app/domain/orders/entities"
	orderRepo "ecommerce-app/domain/orders/repositories"
	productRepo "ecommerce-app/domain/products/repositories"
	"ecommerce-app/config"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func StartInventoryWorker(ctx context.Context, db *gorm.DB, orderRepository orderRepo.OrderRepository, productRepository productRepo.ProductRepository) error {
	ch, err := config.NewChannel()
	if err != nil {
		return err
	}

	exchange := os.Getenv("RABBITMQ_EXCHANGE")
	if exchange == "" {
		exchange = "orders_direct"
	}
	placedRoutingKey := os.Getenv("RABBITMQ_ROUTING_KEY")
	if placedRoutingKey == "" {
		placedRoutingKey = "order.placed"
	}
	placedQueue := os.Getenv("RABBITMQ_QUEUE")
	if placedQueue == "" {
		placedQueue = "order_placed_queue"
	}

	if err := config.EnsureDirectExchange(ch, exchange); err != nil {
		return err
	}
	_, err = config.DeclareQuorumQueue(ch, placedQueue, exchange, placedRoutingKey)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		placedQueue,
		"",    
		false, 
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Inventory worker: consuming", placedQueue)

	go func() {
		for {
			select {
			case <-ctx.Done():
				_ = ch.Close()
				return
			case d, ok := <-msgs:
				if !ok {
					log.Println("Inventory worker: delivery channel closed")
					return
				}
				start := time.Now()

				var payload events.OrderPlacedPayload
				if err := json.Unmarshal(d.Body, &payload); err != nil {
					log.Printf("inventory: invalid message, nack & discard: %v", err)
					d.Nack(false, false)
					continue
				}

				if err := processOrder(db, &payload, orderRepository); err != nil {
					log.Printf("inventory: processing error for order %s: %v", payload.OrderID, err)
					d.Nack(false, true) 
					continue
				}

				d.Ack(false)
				log.Printf("inventory: processed order %s in %s", payload.OrderID, time.Since(start))
			}
		}
	}()

	return nil
}

func processOrder(db *gorm.DB, p *events.OrderPlacedPayload, orderRepository orderRepo.OrderRepository) error {
	return db.Transaction(func(tx *gorm.DB) error {
		order, err := orderRepository.FindByID(p.OrderID)
		if err != nil {
			return err
		}

		if order.Status != "PENDING" {
			return nil
		}

		for _, item := range p.Items {
			var prod prodEntities.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&prod, "id = ?", item.ProductID).Error; err != nil {
				return err
			}
			if prod.Stock < item.Quantity {
				order.Status = "CANCELLED"
				if err := tx.Save(order).Error; err != nil {
					return err
				}
				go publishOrderResult(order, "CANCELLED", "out_of_stock")
				return nil 
			}
			prod.Stock = prod.Stock - item.Quantity
			if err := tx.Save(&prod).Error; err != nil {
				return err
			}
		}

		order.Status = "CONFIRMED"
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		go publishOrderResult(order, "CONFIRMED", "")
		return nil
	})
}

func publishOrderResult(order *orderEntities.Order, status, reason string) {
	ch, err := config.NewChannel()
	if err != nil {
		log.Printf("inventory: publish channel error: %v", err)
		return
	}
	defer ch.Close()

	exchange := os.Getenv("RABBITMQ_EXCHANGE")
	if exchange == "" {
		exchange = "orders_direct"
	}
	confirmedRoutingKey := os.Getenv("RABBITMQ_CONFIRM_ROUTING_KEY")
	if confirmedRoutingKey == "" {
		confirmedRoutingKey = "order.confirmed"
	}
	failedRoutingKey := os.Getenv("RABBITMQ_FAILED_ROUTING_KEY")
	if failedRoutingKey == "" {
		failedRoutingKey = "order.failed"
	}

	payload := events.OrderResultPayload{
		OrderID: order.ID,
		UserID:  order.UserID,
		Status:  status,
		Reason:  reason,
		CreatedAt: time.Now().UTC(),
	}
	body, _ := json.Marshal(payload)

	_ = config.EnsureDirectExchange(ch, exchange)
	_, _ = config.DeclareQuorumQueue(ch, os.Getenv("RABBITMQ_CONFIRM_QUEUE"), exchange, confirmedRoutingKey)
	_, _ = config.DeclareQuorumQueue(ch, os.Getenv("RABBITMQ_FAILED_QUEUE"), exchange, failedRoutingKey)

	rk := confirmedRoutingKey
	if status == "CANCELLED" {
		rk = failedRoutingKey
	}

	if err :=  config.PublishJSON(ch, exchange, rk, body); err != nil {
		log.Printf("inventory: failed publish order result %s: %v", order.ID, err)
		return
	}
	log.Printf("inventory: published order_%s for order %s", status, order.ID)
}
