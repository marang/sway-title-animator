package swayipc

import "fmt"

// CommandOutcomeUnknownError reports a mutating request whose connection
// failed after the exchange began. Callers must observe fresh compositor state
// before deciding whether another command is needed.
type CommandOutcomeUnknownError struct {
	Cause error
}

func (err *CommandOutcomeUnknownError) Error() string {
	return fmt.Sprintf("sway command outcome is unknown after connection failure: %v", err.Cause)
}

func (err *CommandOutcomeUnknownError) Unwrap() error {
	return err.Cause
}

// Client reconnects once for read-only requests. Mutating requests are never
// repeated automatically because their outcome may be ambiguous.
type Client struct {
	socket string
	conn   *Conn
}

// NewClient creates a reconnecting Sway IPC request client.
func NewClient(socket string) *Client {
	return &Client{socket: socket}
}

// Close closes the current control connection, if any.
func (client *Client) Close() {
	if client == nil || client.conn == nil {
		return
	}
	_ = client.conn.Close()
	client.conn = nil
}

// Request sends one request. Known read-only requests reconnect once after a
// failed exchange; RUN_COMMAND returns CommandOutcomeUnknownError instead.
func (client *Client) Request(messageType MessageType, payload []byte) (Message, error) {
	if err := validatePayloadSize(payload); err != nil {
		return Message{}, err
	}
	attempts := 1
	if retryableReadOnlyRequest(messageType) {
		attempts = 2
	}
	var lastErr error
	for range attempts {
		if err := client.ensure(); err != nil {
			return Message{}, err
		}
		message, err := client.conn.Request(messageType, payload)
		if err == nil {
			return message, nil
		}
		lastErr = err
		client.Close()
	}
	if messageType == RunCommand {
		return Message{}, &CommandOutcomeUnknownError{Cause: lastErr}
	}
	if attempts == 1 {
		return Message{}, fmt.Errorf("sway ipc request failed: %w", lastErr)
	}
	return Message{}, fmt.Errorf("sway ipc request failed after reconnect: %w", lastErr)
}

func retryableReadOnlyRequest(messageType MessageType) bool {
	return messageType == GetTree
}

func (client *Client) ensure() error {
	if client == nil {
		return fmt.Errorf("sway ipc client is nil")
	}
	if client.conn != nil {
		return nil
	}
	conn, err := Dial(client.socket)
	if err != nil {
		return err
	}
	client.conn = conn
	return nil
}
