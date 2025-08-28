package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventProcessorConfig defines configuration for downstream event processors
// Defines which events get processed by which downstream systems for specific clients
type EventProcessorConfig struct {
	ID            primitive.ObjectID     `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string                `bson:"name" json:"name" validate:"required"`
	Description   string                `bson:"description,omitempty" json:"description,omitempty"`
	ClientID      primitive.ObjectID     `bson:"client" json:"client_id" validate:"required"`
	ProcessorType ProcessorType         `bson:"processor_type" json:"processor_type" validate:"required"`
	Config        map[string]interface{} `bson:"config" json:"config" validate:"required"` // Processor-specific configuration
	EventTypes    []EventType           `bson:"event_types" json:"event_types"`           // Which events this processor handles
	EntityTypes   []EntityType          `bson:"entity_types" json:"entity_types"`         // Which entity types this processor handles
	IsActive      bool                  `bson:"is_active" json:"is_active"`
	CreatedAt     time.Time             `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time             `bson:"updated_at" json:"updated_at"`
}

// TableName returns the collection name for EventProcessorConfig
func (EventProcessorConfig) TableName() string {
	return "event_processor_configs"
}

// BeforeCreate sets the timestamps before creating
func (epc *EventProcessorConfig) BeforeCreate() {
	now := time.Now().UTC()
	epc.CreatedAt = now
	epc.UpdatedAt = now
	if epc.ID.IsZero() {
		epc.ID = primitive.NewObjectID()
	}
	if epc.Config == nil {
		epc.Config = make(map[string]interface{})
	}
	if epc.EventTypes == nil {
		epc.EventTypes = make([]EventType, 0)
	}
	if epc.EntityTypes == nil {
		epc.EntityTypes = make([]EntityType, 0)
	}
	epc.IsActive = true // Default to active
}

// BeforeUpdate sets the updated timestamp before updating
func (epc *EventProcessorConfig) BeforeUpdate() {
	epc.UpdatedAt = time.Now().UTC()
}

// BaseProcessorConfig represents the base configuration for all processor types.
type BaseProcessorConfig struct {
	// Common fields can be added here if needed
}

// HttpWebhookConfig represents HTTP webhook processor configuration.
type HttpWebhookConfig struct {
	WebhookURL string            `json:"webhook_url" bson:"webhook_url"`
	Headers    map[string]string `json:"headers" bson:"headers"`
	Timeout    int               `json:"timeout" bson:"timeout"` // in seconds
}

// AmqpConfig represents AMQP processor configuration.
type AmqpConfig struct {
	Host       string  `json:"host" bson:"host"`
	Port       int     `json:"port" bson:"port"`
	VHost      string  `json:"vhost" bson:"vhost"`
	Exchange   string  `json:"exchange" bson:"exchange"`
	RoutingKey string  `json:"routing_key" bson:"routing_key"`
	Username   *string `json:"username,omitempty" bson:"username,omitempty"`
	Password   *string `json:"password,omitempty" bson:"password,omitempty"`
}

// GetHttpWebhookConfig extracts HTTP webhook configuration from the config map.
func (epc *EventProcessorConfig) GetHttpWebhookConfig() (*HttpWebhookConfig, error) {
	if epc.ProcessorType != ProcessorTypeHTTPWebhook {
		return nil, fmt.Errorf("processor type is not HTTP_WEBHOOK")
	}

	config := &HttpWebhookConfig{}
	if webhookURL, ok := epc.Config["webhook_url"].(string); ok {
		config.WebhookURL = webhookURL
	}
	if headers, ok := epc.Config["headers"].(map[string]interface{}); ok {
		config.Headers = make(map[string]string)
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				config.Headers[k] = strVal
			}
		}
	}
	if timeout, ok := epc.Config["timeout"].(int); ok {
		config.Timeout = timeout
	} else if timeout, ok := epc.Config["timeout"].(float64); ok {
		config.Timeout = int(timeout)
	} else {
		config.Timeout = 10 // default timeout
	}

	return config, nil
}

// GetAmqpConfig extracts AMQP configuration from the config map.
func (epc *EventProcessorConfig) GetAmqpConfig() (*AmqpConfig, error) {
	if epc.ProcessorType != ProcessorTypeAMQP {
		return nil, fmt.Errorf("processor type is not AMQP")
	}

	config := &AmqpConfig{}
	if host, ok := epc.Config["host"].(string); ok {
		config.Host = host
	}
	if port, ok := epc.Config["port"].(int); ok {
		config.Port = port
	} else if port, ok := epc.Config["port"].(float64); ok {
		config.Port = int(port)
	} else {
		config.Port = 5672 // default port
	}
	if vhost, ok := epc.Config["vhost"].(string); ok {
		config.VHost = vhost
	} else {
		config.VHost = "/" // default vhost
	}
	if exchange, ok := epc.Config["exchange"].(string); ok {
		config.Exchange = exchange
	}
	if routingKey, ok := epc.Config["routing_key"].(string); ok {
		config.RoutingKey = routingKey
	}
	if username, ok := epc.Config["username"].(string); ok {
		config.Username = &username
	}
	if password, ok := epc.Config["password"].(string); ok {
		config.Password = &password
	}

	return config, nil
}

// ValidateConfig validates the config against the appropriate schema based on processor_type
func (epc *EventProcessorConfig) ValidateConfig() error {
	switch epc.ProcessorType {
	case ProcessorTypeHTTPWebhook:
		_, err := epc.GetHttpWebhookConfig()
		return err
	case ProcessorTypeAMQP:
		_, err := epc.GetAmqpConfig()
		return err
	default:
		return fmt.Errorf("unsupported processor type: %s", epc.ProcessorType)
	}
}