import struct


def recv_exact(sock, n):
    """Receive exactly n bytes from socket, avoiding short reads."""
    data = b''
    while len(data) < n:
        chunk = sock.recv(n - len(data))
        if not chunk:
            raise OSError("Connection closed before receiving all data")
        data += chunk
    return data


def recv_message(sock):
    """Receive a length-prefixed message from socket."""
    header = recv_exact(sock, 2)
    length = struct.unpack('!H', header)[0]
    payload = recv_exact(sock, length)
    return payload.decode('utf-8')


def send_message(sock, message):
    """Send a length-prefixed message through socket, avoiding short writes."""
    payload = message.encode('utf-8')
    header = struct.pack('!H', len(payload))
    data = header + payload
    total_sent = 0
    while total_sent < len(data):
        sent = sock.send(data[total_sent:])
        if sent == 0:
            raise OSError("Connection closed before sending all data")
        total_sent += sent