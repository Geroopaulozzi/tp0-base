package common

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

type Client struct {
	config ClientConfig
	conn   net.Conn
}

func NewClient(config ClientConfig) *Client {
	return &Client{config: config}
}

func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

func (c *Client) StartClientLoop() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM)

	nombre := os.Getenv("NOMBRE")
	apellido := os.Getenv("APELLIDO")
	documento := os.Getenv("DOCUMENTO")
	nacimiento := os.Getenv("NACIMIENTO")
	numero := os.Getenv("NUMERO")

	select {
	case <-sigChan:
		c.shutdown()
		return
	default:
	}

	if err := c.createClientSocket(); err != nil {
		return
	}

	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		c.config.ID, nombre, apellido, documento, nacimiento, numero,
	)

	if err := sendMessage(c.conn, payload); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		c.conn.Close()
		return
	}

	response, err := recvMessage(c.conn)
	c.conn.Close()
	log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)

	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}

	if response == "OK" {
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			documento, numero)
	} else {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | response: %v",
			c.config.ID, response)
	}

	select {
	case <-sigChan:
		c.shutdown()
	case <-time.After(c.config.LoopPeriod):
	}
}

func (c *Client) shutdown() {
	log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
	log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
}