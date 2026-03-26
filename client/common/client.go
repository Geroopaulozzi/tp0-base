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

	// Una sola conexión para todo el flujo
	if err := c.createClientSocket(); err != nil {
		return
	}
	defer func() {
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}()

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
			fields := strings.Split(line, ",")
			if len(fields) != 5 {
				log.Errorf("action: parse_bet | result: fail | client_id: %v | line: %v",
					c.config.ID, line)
				continue
			}
			bet := fmt.Sprintf("%s|%s|%s|%s|%s|%s",
				c.config.ID, fields[0], fields[1], fields[2], fields[3], fields[4])
			batch = append(batch, bet)
		}

		if len(batch) == c.config.BatchMaxAmount || (!hasLine && len(batch) > 0) {
			if err := c.sendBatch(batch); err != nil {
				return
			}
			batch = batch[:0]
		}

		if !hasLine {
			break
		}
	}

	select {
	case <-sigChan:
		c.shutdown()
		return
	default:
	}

	if err := c.sendFin(); err != nil {
		return
	}

	select {
	case <-sigChan:
		c.shutdown()
		return
	default:
	}

	c.consultarGanadores(sigChan)
}

func (c *Client) sendBatch(batch []string) error {
	payload := "B|" + strings.Join(batch, "\n")

	if err := sendMessage(c.conn, payload); err != nil {
		log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	response, err := recvMessage(c.conn)
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

func (c *Client) sendFin() error {
	if err := sendMessage(c.conn, "F"); err != nil {
		log.Errorf("action: fin_envio | result: fail | client_id: %v | error: %v",
			c.config.ID, err)
		return err
	}

	log.Infof("action: fin_envio | result: success | client_id: %v", c.config.ID)
	return nil
}

func (c *Client) consultarGanadores(sigChan chan os.Signal) {
	for {
		select {
		case <-sigChan:
			c.shutdown()
			return
		default:
		}

		if err := sendMessage(c.conn, "G"+c.config.ID); err != nil {
			log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return
		}

		response, err := recvMessage(c.conn)
		if err != nil {
			log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v",
				c.config.ID, err)
			return
		}

		if response == "NOT_READY" {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		cant := 0
		if response != "" {
			winners := strings.Split(strings.TrimSpace(response), "\n")
			for _, w := range winners {
				if strings.TrimSpace(w) != "" {
					cant++
				}
			}
		}

		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", cant)
		return
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