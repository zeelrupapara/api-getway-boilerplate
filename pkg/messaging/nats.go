package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/google/uuid"

	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/config"
	"gitlab.com/flexgrewtechnologies/greenlync-api-gateway/pkg/logger"
)

// NATS represents the NATS messaging client
type NATS struct {
	Conn   *nats.Conn
	Config *config.NATSConfig
	Logger *logger.Logger
}

// Message represents a structured message for cannabis platform
type Message struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Subject     string                 `json:"subject"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	
	// Cannabis compliance context
	UserID       string `json:"user_id,omitempty"`
	DispensaryID string `json:"dispensary_id,omitempty"`
	Compliance   bool   `json:"compliance,omitempty"`
	
	// Message metadata
	Priority     int    `json:"priority"`
	TTL          int64  `json:"ttl,omitempty"`
	Retry        int    `json:"retry,omitempty"`
}

// Cannabis platform subjects
const (
	// User events
	SubjectUserCreated     = "greenlync.user.created"
	SubjectUserUpdated     = "greenlync.user.updated"
	SubjectUserDeleted     = "greenlync.user.deleted"
	SubjectUserLogin       = "greenlync.user.login"
	SubjectUserLogout      = "greenlync.user.logout"
	
	// Cannabis compliance events
	SubjectAgeVerified     = "greenlync.compliance.age_verified"
	SubjectStateVerified   = "greenlync.compliance.state_verified"
	SubjectComplianceCheck = "greenlync.compliance.check"
	
	// Order events
	SubjectOrderCreated    = "greenlync.order.created"
	SubjectOrderUpdated    = "greenlync.order.updated"
	SubjectOrderCompleted  = "greenlync.order.completed"
	SubjectOrderCancelled  = "greenlync.order.cancelled"
	
	// Product events
	SubjectProductCreated  = "greenlync.product.created"
	SubjectProductUpdated  = "greenlync.product.updated"
	SubjectInventoryUpdate = "greenlync.inventory.updated"
	
	// Payment events
	SubjectPaymentCreated  = "greenlync.payment.created"
	SubjectPaymentCompleted = "greenlync.payment.completed"
	SubjectPaymentFailed   = "greenlync.payment.failed"
	
	// Dispensary events
	SubjectDispensaryCreated = "greenlync.dispensary.created"
	SubjectDispensaryUpdated = "greenlync.dispensary.updated"
	
	// WebSocket events
	SubjectWSConnected     = "greenlync.ws.connected"
	SubjectWSDisconnected  = "greenlync.ws.disconnected"
	SubjectWSMessage       = "greenlync.ws.message"
	
	// System events
	SubjectSystemAlert     = "greenlync.system.alert"
	SubjectHealthCheck     = "greenlync.system.health"
	
	// Audit events
	SubjectAuditLog        = "greenlync.audit.log"
	SubjectComplianceAudit = "greenlync.audit.compliance"
)

// NewNATS creates a new NATS client connection
func NewNATS(cfg *config.NATSConfig, log *logger.Logger) (*NATS, error) {
	log.Info("Initializing NATS connection",
		"url", cfg.URL,
		"cluster_id", cfg.ClusterID,
		"client_id", cfg.ClientID,
	)

	// Configure NATS options
	opts := []nats.Option{
		nats.Name(cfg.ClientID),
		nats.Timeout(cfg.ConnectTimeout),
		nats.ReconnectWait(cfg.ReconnectWait),
		nats.MaxReconnects(cfg.MaxReconnects),
		nats.PingInterval(cfg.PingInterval),
		nats.MaxPingsOutstanding(cfg.MaxPingsOut),
		
		// Callbacks for connection events
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			log.Warn("NATS disconnected", "error", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Info("NATS reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Info("NATS connection closed")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			log.Error("NATS error", "error", err, "subject", sub.Subject)
		}),
	}

	// Connect to NATS
	conn, err := nats.Connect(cfg.URL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	natsClient := &NATS{
		Conn:   conn,
		Config: cfg,
		Logger: log,
	}

	log.Info("NATS connection established successfully",
		"connected_url", conn.ConnectedUrl(),
		"server_info", conn.ConnectedServerName(),
	)

	return natsClient, nil
}

// Close closes the NATS connection
func (n *NATS) Close() error {
	if n.Conn != nil && !n.Conn.IsClosed() {
		n.Conn.Close()
		n.Logger.Info("NATS connection closed successfully")
	}
	return nil
}

// Health checks the NATS connection health
func (n *NATS) Health() error {
	if n.Conn == nil || n.Conn.IsClosed() {
		return fmt.Errorf("NATS connection is closed")
	}
	
	if !n.Conn.IsConnected() {
		return fmt.Errorf("NATS not connected")
	}
	
	return nil
}

// GetStats returns NATS connection statistics
func (n *NATS) GetStats() map[string]interface{} {
	if n.Conn == nil || n.Conn.IsClosed() {
		return map[string]interface{}{
			"status": "disconnected",
		}
	}

	stats := n.Conn.Statistics
	return map[string]interface{}{
		"status":           "connected",
		"connected_url":    n.Conn.ConnectedUrl(),
		"server_name":      n.Conn.ConnectedServerName(),
		"in_msgs":          stats.InMsgs,
		"out_msgs":         stats.OutMsgs,
		"in_bytes":         stats.InBytes,
		"out_bytes":        stats.OutBytes,
		"reconnects":       stats.Reconnects,
		"outstanding_reqs": stats.OutstandingRequests,
	}
}

// Publishing methods

// Publish publishes a message to a subject
func (n *NATS) Publish(subject string, data interface{}) error {
	message := n.createMessage(subject, "publish", data)
	return n.publishMessage(subject, message)
}

// PublishCannabisEvent publishes a cannabis-specific event with compliance context
func (n *NATS) PublishCannabisEvent(subject, userID, dispensaryID string, data interface{}, compliance bool) error {
	message := n.createMessage(subject, "cannabis_event", data)
	message.UserID = userID
	message.DispensaryID = dispensaryID
	message.Compliance = compliance
	message.Priority = 1 // High priority for compliance events
	
	// Log cannabis event for audit trail
	n.Logger.LogCannabisAudit(userID, "event_published", subject, map[string]interface{}{
		"message_id":    message.ID,
		"dispensary_id": dispensaryID,
		"compliance":    compliance,
		"data":         data,
	})
	
	return n.publishMessage(subject, message)
}

// PublishUserEvent publishes user-related events
func (n *NATS) PublishUserEvent(eventType, userID, dispensaryID string, data interface{}) error {
	subject := fmt.Sprintf("greenlync.user.%s", eventType)
	return n.PublishCannabisEvent(subject, userID, dispensaryID, data, true)
}

// PublishOrderEvent publishes order-related events
func (n *NATS) PublishOrderEvent(eventType, userID, dispensaryID, orderID string, data interface{}) error {
	subject := fmt.Sprintf("greenlync.order.%s", eventType)
	orderData := map[string]interface{}{
		"order_id": orderID,
		"data":    data,
	}
	return n.PublishCannabisEvent(subject, userID, dispensaryID, orderData, true)
}

// PublishComplianceEvent publishes compliance-related events
func (n *NATS) PublishComplianceEvent(eventType, userID, dispensaryID string, data interface{}) error {
	subject := fmt.Sprintf("greenlync.compliance.%s", eventType)
	return n.PublishCannabisEvent(subject, userID, dispensaryID, data, true)
}

// PublishSystemEvent publishes system-level events
func (n *NATS) PublishSystemEvent(eventType string, data interface{}) error {
	subject := fmt.Sprintf("greenlync.system.%s", eventType)
	message := n.createMessage(subject, "system_event", data)
	message.Priority = 2 // Medium priority for system events
	
	return n.publishMessage(subject, message)
}

// PublishAuditEvent publishes audit events for regulatory compliance
func (n *NATS) PublishAuditEvent(userID, action, resource string, data interface{}) error {
	auditData := map[string]interface{}{
		"action":    action,
		"resource":  resource,
		"data":     data,
		"timestamp": time.Now().UTC(),
	}
	
	return n.PublishCannabisEvent(SubjectAuditLog, userID, "", auditData, true)
}

// Subscription methods

// Subscribe subscribes to a subject with a handler function
func (n *NATS) Subscribe(subject string, handler func(*Message)) (*nats.Subscription, error) {
	sub, err := n.Conn.Subscribe(subject, func(msg *nats.Msg) {
		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			n.Logger.Error("Failed to unmarshal NATS message",
				"error", err,
				"subject", msg.Subject,
			)
			return
		}
		
		// Log message processing for cannabis compliance
		if message.Compliance {
			n.Logger.LogCannabisAudit(message.UserID, "message_processed", msg.Subject, map[string]interface{}{
				"message_id":    message.ID,
				"dispensary_id": message.DispensaryID,
				"type":         message.Type,
			})
		}
		
		handler(&message)
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to subject %s: %w", subject, err)
	}
	
	n.Logger.Info("Subscribed to NATS subject", "subject", subject)
	return sub, nil
}

// SubscribeQueue subscribes to a subject with queue group for load balancing
func (n *NATS) SubscribeQueue(subject, queue string, handler func(*Message)) (*nats.Subscription, error) {
	sub, err := n.Conn.QueueSubscribe(subject, queue, func(msg *nats.Msg) {
		var message Message
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			n.Logger.Error("Failed to unmarshal NATS message",
				"error", err,
				"subject", msg.Subject,
				"queue", queue,
			)
			return
		}
		
		// Log message processing for cannabis compliance
		if message.Compliance {
			n.Logger.LogCannabisAudit(message.UserID, "queue_message_processed", msg.Subject, map[string]interface{}{
				"message_id":    message.ID,
				"dispensary_id": message.DispensaryID,
				"queue":        queue,
				"type":         message.Type,
			})
		}
		
		handler(&message)
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to queue %s on subject %s: %w", queue, subject, err)
	}
	
	n.Logger.Info("Subscribed to NATS subject with queue", "subject", subject, "queue", queue)
	return sub, nil
}

// Request/Reply methods

// Request sends a request and waits for a reply
func (n *NATS) Request(subject string, data interface{}, timeout time.Duration) (*Message, error) {
	message := n.createMessage(subject, "request", data)
	
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request message: %w", err)
	}
	
	reply, err := n.Conn.Request(subject, msgBytes, timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	
	var replyMessage Message
	if err := json.Unmarshal(reply.Data, &replyMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reply message: %w", err)
	}
	
	return &replyMessage, nil
}

// Helper methods

// createMessage creates a new structured message
func (n *NATS) createMessage(subject, msgType string, data interface{}) *Message {
	return &Message{
		ID:        uuid.New().String(),
		Type:      msgType,
		Source:    "greenlync-api-gateway",
		Subject:   subject,
		Data:      n.convertToMap(data),
		Timestamp: time.Now().UTC(),
		Priority:  3, // Default priority
	}
}

// publishMessage publishes a structured message
func (n *NATS) publishMessage(subject string, message *Message) error {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	if err := n.Conn.Publish(subject, msgBytes); err != nil {
		return fmt.Errorf("failed to publish message to %s: %w", subject, err)
	}
	
	n.Logger.Debug("Published NATS message",
		"subject", subject,
		"message_id", message.ID,
		"type", message.Type,
		"priority", message.Priority,
	)
	
	return nil
}

// convertToMap converts interface{} to map[string]interface{}
func (n *NATS) convertToMap(data interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}
	
	// If already a map, return as is
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}
	
	// Convert via JSON marshaling/unmarshaling
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		n.Logger.Warn("Failed to marshal data to map", "error", err)
		return map[string]interface{}{
			"raw_data": fmt.Sprintf("%+v", data),
		}
	}
	
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		n.Logger.Warn("Failed to unmarshal data to map", "error", err)
		return map[string]interface{}{
			"raw_data": string(jsonBytes),
		}
	}
	
	return result
}

// Cannabis-specific utility functions

// SetupCannabisSubscriptions sets up default subscriptions for cannabis platform
func (n *NATS) SetupCannabisSubscriptions() error {
	n.Logger.Info("Setting up cannabis platform NATS subscriptions...")
	
	// Subscribe to compliance events
	if _, err := n.Subscribe(SubjectComplianceCheck, n.handleComplianceEvent); err != nil {
		return fmt.Errorf("failed to subscribe to compliance events: %w", err)
	}
	
	// Subscribe to audit events
	if _, err := n.SubscribeQueue(SubjectAuditLog, "audit-workers", n.handleAuditEvent); err != nil {
		return fmt.Errorf("failed to subscribe to audit events: %w", err)
	}
	
	// Subscribe to system health checks
	if _, err := n.Subscribe(SubjectHealthCheck, n.handleHealthCheck); err != nil {
		return fmt.Errorf("failed to subscribe to health check events: %w", err)
	}
	
	n.Logger.Info("Cannabis platform NATS subscriptions setup completed")
	return nil
}

// Default event handlers

// handleComplianceEvent handles cannabis compliance events
func (n *NATS) handleComplianceEvent(msg *Message) {
	n.Logger.LogCannabisAudit(msg.UserID, "compliance_event_received", msg.Subject, map[string]interface{}{
		"message_id":    msg.ID,
		"dispensary_id": msg.DispensaryID,
		"data":         msg.Data,
	})
}

// handleAuditEvent handles audit events for regulatory compliance
func (n *NATS) handleAuditEvent(msg *Message) {
	n.Logger.Info("Processing audit event",
		"message_id", msg.ID,
		"user_id", msg.UserID,
		"dispensary_id", msg.DispensaryID,
		"data", msg.Data,
	)
}

// handleHealthCheck handles system health check events
func (n *NATS) handleHealthCheck(msg *Message) {
	n.Logger.Debug("Received health check event",
		"message_id", msg.ID,
		"source", msg.Source,
		"data", msg.Data,
	)
}