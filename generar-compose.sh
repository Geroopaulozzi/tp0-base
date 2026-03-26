#!/bin/bash

if [ "$#" -ne 2 ]; then
  echo "Uso: $0 <archivo_salida> <cantidad_clientes>"
  exit 1
fi

ARCHIVO_SALIDA="$1"
CANT_CLIENTES="$2"

if ! [[ "$CANT_CLIENTES" =~ ^[0-9]+$ ]] || [ "$CANT_CLIENTES" -lt 0 ]; then
  echo "Error: la cantidad de clientes debe ser un entero positivo."
  exit 1
fi

cat > "$ARCHIVO_SALIDA" <<EOF
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
    volumes:
      - ./server/config.ini:/config.ini
    networks:
      - testing_net

EOF

for i in $(seq 1 "$CANT_CLIENTES"); do
cat >> "$ARCHIVO_SALIDA" <<EOF
  client$i:
    container_name: client$i
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=$i
      - NOMBRE=Geronimo
      - APELLIDO=Paulozzi
      - DOCUMENTO=41834183$i
      - NACIMIENTO=1999-04-29
      - NUMERO=7574
    volumes:
      - ./client/config.yaml:/config.yaml
    networks:
      - testing_net
    depends_on:
      - server

EOF
done

cat >> "$ARCHIVO_SALIDA" <<EOF
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
EOF

echo "Archivo generado: $ARCHIVO_SALIDA con $CANT_CLIENTES clientes."