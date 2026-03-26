import socket
import logging
from common.protocol import recv_message, send_message
from common.utils import Bet, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True

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
            bets = []
            for line in msg.split('\n'):
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
            try:
                send_message(client_sock, 'ERROR')
            except OSError:
                pass
        finally:
            client_sock.close()
            logging.info("action: close_client_socket | result: success")

    def __accept_new_connection(self):
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c