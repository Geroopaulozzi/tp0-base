# Resolución TP0

## Ejercicio 1

### Objetivo
El objetivo de este ejercicio fue generar automáticamente un archivo de Docker Compose con una cantidad configurable de clientes, en lugar de mantener una definición fija con un único cliente.

La consigna pide un script bash ubicado en la raíz del proyecto, que reciba como parámetros el nombre del archivo de salida y la cantidad de clientes a generar.

```bash
./generar-compose.sh <archivo_salida> <cantidad_clientes>
```

### Cambios realizados
Para resolver este ejercicio se agregó el archivo `generar-compose.sh` en la raíz del proyecto.

Este script:
1. valida que se reciban exactamente dos parámetros
2. valida que la cantidad de clientes sea un entero positivo
3. genera el bloque correspondiente al `server`
4. genera dinámicamente los clientes `client1`, `client2`, ..., `clientN`
5. asigna correctamente la variable de entorno `CLI_ID` según el número de cliente
6. escribe también la definición de la red `testing_net`

### Decisión de implementación
Se decidió resolver este ejercicio únicamente con bash, sin usar scripts auxiliares en Python o Go, ya que la generación del archivo YAML requerido podía resolverse de forma simple con redirecciones y un bucle `for`.

Se tomó como base el archivo `docker-compose-dev.yaml` provisto originalmente en el proyecto, manteniendo:
- el servicio `server`
- la configuración de red
- las imágenes utilizadas
- las variables de entorno ya definidas

La parte variable del problema fue la generación de los clientes.

### Funcionamiento
El script recibe un nombre de archivo de salida y una cantidad de clientes.

```bash
./generar-compose.sh <archivo_salida> <cantidad_clientes>
```

A partir de eso, genera un archivo Docker Compose equivalente al original, pero con una cantidad variable de clientes.

Por ejemplo, al ejecutar el script con salida `docker-compose-dev.yaml` y cantidad `5`, se obtiene un archivo con:
- 1 servicio `server`
- 5 servicios cliente: `client1`, `client2`, `client3`, `client4` y `client5`

Cada cliente mantiene la misma estructura base, cambiando:
- el nombre del servicio
- `container_name`
- la variable `CLI_ID`

### Validaciones incorporadas
Se agregaron dos validaciones:
- que el script reciba exactamente dos argumentos
- que la cantidad de clientes sea un número entero positivo

En caso contrario, el script finaliza mostrando un mensaje de error.

### Verificación realizada
Para comprobar que la solución funcionaba correctamente se hicieron dos verificaciones.

Primero, la generación del archivo.

```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

Después, la validación del YAML generado con Docker Compose.

```bash
docker compose -f docker-compose-dev.yaml config
```

El comando de validación fue exitoso, por lo que se verificó que el archivo generado es sintácticamente válido para Docker Compose.

## Ejercicio 2

### Objetivo
El objetivo de este ejercicio fue modificar el cliente y el servidor para que los cambios en los archivos de configuración no requieran reconstruir las imágenes de Docker para ser efectivos.

La consigna pide que la configuración de cada componente sea inyectada en el contenedor y permanezca fuera de la imagen, utilizando volúmenes.

### Situación inicial
Antes de los cambios, tanto el cliente como el servidor dependían de archivos de configuración que quedaban embebidos dentro de la imagen Docker durante el proceso de build.

En el caso del cliente, el `Dockerfile` copiaba explícitamente `config.yaml` a la imagen final.

En el caso del servidor, el `Dockerfile` copiaba todo el directorio `server/`, lo que también incluía `config.ini`.

Esto implicaba que, ante cualquier modificación en los archivos de configuración, fuera necesario reconstruir la imagen para que los cambios impactaran en la ejecución.

### Cambios realizados
Para resolver este ejercicio se hicieron tres cambios principales:

1. se modificó el `Dockerfile` del cliente para dejar de copiar `config.yaml` dentro de la imagen
2. se modificó el `Dockerfile` del servidor para dejar de copiar `config.ini` dentro de la imagen
3. se actualizó `generar-compose.sh` para que el archivo `docker-compose` generado monte ambos archivos de configuración como volúmenes

### Cambios en el cliente
En el cliente se eliminó la copia del archivo `config.yaml` dentro de la imagen final.

De esta forma, la imagen pasó a contener únicamente el binario necesario para ejecutar la aplicación, mientras que la configuración quedó desacoplada del proceso de build.

Luego, en el archivo de Docker Compose generado, se agregó el montaje del archivo local `./client/config.yaml` dentro del contenedor en la ruta `/config.yaml`.

### Cambios en el servidor
En el servidor se dejó de copiar el directorio completo `server/` a la imagen, ya que eso incluía también el archivo `config.ini`.

En su lugar, se copiaron únicamente los archivos y directorios necesarios para ejecutar la aplicación y correr los tests, excluyendo explícitamente `config.ini`.

Luego, en el archivo de Docker Compose generado, se agregó el montaje del archivo local `./server/config.ini` dentro del contenedor en la ruta `/config.ini`.

### Cambios en `generar-compose.sh`
Además de generar una cantidad variable de clientes, el script pasó a incluir la definición de volúmenes para ambos servicios.

En el caso del servidor, se monta `./server/config.ini` sobre `/config.ini`.

En el caso de cada cliente, se monta `./client/config.yaml` sobre `/config.yaml`.

Esto permite que cualquier modificación en los archivos de configuración del proyecto se refleje dentro del contenedor sin necesidad de reconstruir la imagen.

### Validación
La primera validación consistió en regenerar el archivo `docker-compose-dev.yaml` y verificar que Docker Compose aceptara correctamente la definición.
El resultado fue exitoso: Docker interpretó correctamente los volúmenes como bind mounts, tanto para el servidor como para cada cliente.

Para comprobar que los cambios en la configuración impactaban sin rebuild, se realizó una prueba práctica en dos etapas.
Primero se levantó el sistema con la configuración inicial y se tomaron los logs como referencia.

```bash
make docker-compose-down
make docker-compose-up
make docker-compose-logs
```

Luego se modificaron valores visibles en los archivos de configuración:
- en `server/config.ini` se cambió `SERVER_LISTEN_BACKLOG` de `5` a `9`
- en `client/config.yaml` se cambió `loop.period` de `5s` a `3s`

Después de eso, los contenedores se volvieron a levantar sin reconstruir las imágenes.

```bash
make docker-compose-down
docker compose -f docker-compose-dev.yaml up -d
```
Finalmente se inspeccionaron los logs nuevamente.

```bash
docker compose -f docker-compose-dev.yaml logs -f
```
### Resultado observado
En los logs del servidor se observó que el valor de `listen_backlog` pasó a ser `9`, confirmando que el contenedor estaba utilizando el nuevo contenido de `server/config.ini`.

En los logs del cliente se observó que `loop_period` pasó a `3s`, confirmando que el contenedor estaba utilizando el nuevo contenido de `client/config.yaml`.

Como esos cambios se reflejaron sin volver a construir las imágenes, quedó validado que la configuración fue correctamente externalizada e inyectada mediante volúmenes.


### Resultado final
Con estos cambios, tanto el cliente como el servidor pasaron a leer sus archivos de configuración desde volúmenes montados en tiempo de ejecución.
De esta manera, el proyecto dejó de requerir una reconstrucción de imágenes ante cambios en `config.ini` y `config.yaml`, cumpliendo con el objetivo del ejercicio 2.

## Ejercicio 3

### Objetivo
El objetivo de este ejercicio fue crear un script de bash que valide el correcto funcionamiento
del servidor utilizando `netcat`, sin instalarlo en el host y sin exponer puertos del servidor.

### Implementación
Se agregó el archivo `validar-echo-server.sh` en la raíz del proyecto.

La restricción central del ejercicio es que netcat no puede instalarse en el host y no se pueden exponer puertos. La solución consiste en levantar un contenedor efímero conectado a la misma red Docker que el servidor, desde el cual se ejecuta netcat internamente.

Docker Compose crea una red llamada `tp0_testing_net` (resultado de combinar el nombre del
proyecto `tp0` con el nombre de la red `testing_net`). Desde cualquier contenedor conectado
a esa red, el servidor es accesible mediante el hostname `server` en el puerto `12345`, sin
necesidad de exponer ese puerto al host.

El script utiliza `docker run --rm` con la imagen `alpine` (que incluye `nc` de forma nativa)
para levantar ese contenedor efímero, enviarle un mensaje al servidor y capturar la respuesta.
```bash
RESPONSE=$(echo "$MESSAGE" | docker run --rm -i \
  --network "$NETWORK" \
  alpine \
  sh -c "nc -w 2 $SERVER $PORT")
```
Dado que el servidor es un echo server, la respuesta esperada es idéntica al mensaje enviado.
El servidor agrega un `\n` al final de cada respuesta, pero la sustitución de comandos en bash
elimina automáticamente el newline final, por lo que la comparación es directa.

Si la respuesta coincide con el mensaje enviado, el script imprime:

```
action: test_echo_server | result: success
```
En caso contrario:

```
action: test_echo_server | result: fail
```

### Verificación
Para verificar el funcionamiento se levantó el sistema normalmente y luego se ejecutó el script.
```bash
make docker-compose-up
chmod +x validar-echo-server.sh
./validar-echo-server.sh
```

El resultado fue `action: test_echo_server | result: success`, confirmando que el servidor
respondió correctamente al mensaje enviado desde el contenedor efímero.

## Ejercicio 4

### Objetivo
El objetivo de este ejercicio fue modificar el cliente y el servidor para que ambos terminen
de forma graceful al recibir la señal SIGTERM, cerrando todos los file descriptors abiertos
y logueando cada cierre antes de que el proceso principal termine.

### Servidor (Python)

Se realizaron cambios en `server/main.py` y `server/common/server.py`.

En `main.py` se registró un handler para SIGTERM usando el módulo `signal`:
```python
signal.signal(signal.SIGTERM, lambda sig, frame: server.stop())
```

En `server.py` se agregó el método `stop()`, que setea una flag `_running = False` y cierra
el server socket. Cerrar el socket hace que la llamada bloqueante `accept()` lance una
`OSError`, lo que permite salir del loop sin necesidad de esperar una nueva conexión.

El cierre del server socket y de cada client socket se loguea explícitamente antes de
que el proceso termine.

### Cliente (Go)

Se modificó `client/common/client.go`.

Se creó un canal que escucha SIGTERM usando `os/signal`:
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGTERM)
```

Se agregaron dos puntos de chequeo con `select` dentro del loop:
- Uno al inicio de cada iteración, antes de abrir la conexión.
- Uno en reemplazo del `time.Sleep`, de forma que si llega SIGTERM durante la espera
  entre mensajes, el cliente responde de inmediato sin tener que esperar a que el
  sleep termine.

En ambos casos se llama a `shutdown()`, que cierra la conexión abierta si la hay y
loguea el cierre antes de terminar.

### Verificación
Para verificar el comportamiento se levantó el sistema y se ejecutó `docker compose stop -t 1`
mientras los clientes todavía estaban corriendo su loop. Luego, antes de ejecutar el down,
se inspeccionaron los logs con `docker compose logs`.

En los logs de cada cliente se observaron los mensajes de shutdown:
```
action: shutdown | result: in_progress | client_id: N
action: close_connection | result: success | client_id: N
action: shutdown | result: success | client_id: N
```

En los logs del servidor se observó que, estando bloqueado en `accept_connections | result: in_progress`,
el SIGTERM disparó el cierre del socket:
```
action: close_server_socket | result: success
```

Esto confirmó que ambos procesos liberaron sus recursos correctamente antes de terminar.

## Ejercicio 5
 
### Objetivo
El objetivo de este ejercicio fue modificar el cliente y el servidor para implementar el caso de uso de la Lotería Nacional. El cliente emula una agencia de quiniela que envía una apuesta al servidor, y el servidor la persiste usando `store_bets(...)`.
 
### Protocolo de comunicación
 
Se implementó un protocolo de mensajes con **longitud prefijada**. Cada mensaje tiene la siguiente estructura:
 
```
[2 bytes: largo del payload (big-endian)][payload UTF-8]
```
 
El header de 2 bytes indica exactamente cuántos bytes conforman el payload, lo que permite que tanto el emisor como el receptor sepan cuánto leer o escribir. Esto evita los fenómenos de short read y short write.
 
El payload de una apuesta tiene el siguiente formato:
 
```
agency|first_name|last_name|document|birthdate|number
```
 
Los campos están separados por `|`. El servidor responde con el mensaje `OK` en caso de éxito, usando el mismo formato de longitud prefijada.
 
### Separación de responsabilidades
 
Se agregaron dos módulos de protocolo, uno por componente:
 
- `server/common/protocol.py`: expone `recv_message` y `send_message`, encapsulando el manejo del header y la lectura/escritura exacta de bytes.
- `client/common/protocol.go`: expone `recvMessage` y `sendMessage` con la misma lógica en Go.
 
De esta forma, `server.py` y `client.go` solo manejan lógica de negocio, delegando la serialización y el transporte al módulo de protocolo.
 
### Cliente (Go)
 
El cliente lee los datos de la apuesta desde variables de entorno: `NOMBRE`, `APELLIDO`, `DOCUMENTO`, `NACIMIENTO` y `NUMERO`. El agency ID se toma del campo `CLI_ID` ya existente en la configuración.
 
Serializa los campos en el formato acordado, envía el mensaje al servidor y espera la confirmación. Al recibirla loguea:
 
```
action: apuesta_enviada | result: success | dni: ${DOCUMENTO} | numero: ${NUMERO}
```
 
### Servidor (Python)
 
El servidor recibe el mensaje, deserializa los campos, construye un objeto `Bet` y llama a `store_bets(...)`. Luego responde con `OK` y loguea:
 
```
action: apuesta_almacenada | result: success | dni: ${document} | numero: ${number}
```
 
### Actualización de `generar-compose.sh`
 
Se agregaron las variables de entorno de la apuesta (`NOMBRE`, `APELLIDO`, `DOCUMENTO`, `NACIMIENTO`, `NUMERO`) en la definición de cada cliente generado. El `DOCUMENTO` varía por cliente para distinguir las apuestas.