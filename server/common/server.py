import socket
import logging
from common.protocol import recv_message, send_message
from common.utils import Bet, store_bets, load_bets, has_won


class Server:
    def __init__(self, port, listen_backlog, total_agencies):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

        self._total_agencies = total_agencies
        self._agencies_done = 0
        self._sorteo_done = False

    def stop(self):
        self._running = False
        self._server_socket.close()
        logging.info("action: close_server_socket | result: success")

    def run(self):
        while self._running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError:
                break

    def __handle_client_connection(self, client_sock):
        try:
            msg = recv_message(client_sock)

            if msg.startswith("B|"):
                self._handle_bets(client_sock, msg[2:])
            elif msg.startswith("F"):
                self._handle_fin(client_sock)
            elif msg.startswith("G"):
                agency_id = msg[1:]
                self._handle_ganadores(client_sock, agency_id)
            else:
                logging.error("action: receive_message | result: fail | reason: unknown message type")
                send_message(client_sock, 'ERROR')

        except (OSError, ValueError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()
            logging.info("action: close_client_socket | result: success")

    def _handle_bets(self, client_sock, payload):
        try:
            bets = []
            for line in payload.split('\n'):
                line = line.strip()
                if not line:
                    continue
                fields = line.split('|')
                if len(fields) != 6:
                    raise ValueError(f"Invalid bet format: {line}")
                agency, first_name, last_name, document, birthdate, number = fields
                bets.append(Bet(agency, first_name, last_name, document, birthdate, number))

            store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
            send_message(client_sock, 'OK')
        except (OSError, ValueError) as e:
            logging.error(f"action: apuesta_recibida | result: fail | cantidad: 0 | error: {e}")
            send_message(client_sock, 'ERROR')

    def _handle_fin(self, client_sock):
        self._agencies_done += 1
        logging.info(f'action: fin_envio | result: success | agencies_done: {self._agencies_done}')

        if self._agencies_done == self._total_agencies:
            self._sorteo_done = True
            logging.info('action: sorteo | result: success')

    def _handle_ganadores(self, client_sock, agency_done):
        if not self._sorteo_done:
            send_message(client_sock, 'NOT_READY')
            return

        winners = []
        for bet in load_bets():
            if bet.agency == int(agency_done) and has_won(bet):
                winners.append(bet.document)

        response = "\n".join(winners)
        send_message(client_sock, response)
        logging.info(f'action: ganadores_enviados | result: success | agency: {agency_done} | cantidad: {len(winners)}')

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c