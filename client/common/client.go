package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

type ClientConfig struct {
	ID             string
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
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

	agencyFile := fmt.Sprintf("/data/agency-%s.csv", c.config.ID)
	file, err := os.Open(agencyFile)
	if err != nil {
		log.Criticalf("action: open_file | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	batch := make([]string, 0, c.config.BatchMaxAmount)

	for {
		select {
		case <-sigChan:
			c.shutdown()
			return
		default:
		}

		hasLine := scanner.Scan()

		if hasLine {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			// CSV format: first_name,last_name,document,birthdate,number
			fields := strings.Split(line, ",")
			if len(fields) != 5 {
				log.Errorf("action: parse_bet | result: fail | client_id: %v | line: %v",
					c.config.ID, line)
				continue
			}
			// Serialize as: agency|first_name|last_name|document|birthdate|number
			bet := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
				c.config.ID, fields[0], fields[1], fields[2], fields[3], fields[4])
			batch = append(batch, bet)
		}

		// Send batch if full or if EOF and batch has content
		if len(batch) == c.config.BatchMaxAmount || (!hasLine && len(batch) > 0) {
			if err := c.sendBatch(batch); err != nil {
				return
			}
			batch = batch[:0]
		}

		if !hasLine {
			break
		}

		select {
		case <-sigChan:
			c.shutdown()
			return
		default:
		}
	}

	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func (c *Client) sendBatch(batch []string) error {
	if err := c.createClientSocket(); err != nil {
		return err
	}

	payload := strings.Join(batch, "\n")

	if err := sendMessage(c.conn, payload); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		c.conn.Close()
		return err
	}

	response, err := recvMessage(c.conn)
	c.conn.Close()
	log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)

	if err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	if response == "OK" {
		log.Infof("action: apuesta_enviada | result: success | client_id: %v | cantidad: %v",
			c.config.ID, len(batch))
	} else {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | response: %v",
			c.config.ID, response)
	}

	return nil
}

func (c *Client) shutdown() {
	log.Infof("action: shutdown | result: in_progress | client_id: %v", c.config.ID)
	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
	log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
}