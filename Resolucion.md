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

## Ejercicio 6

### Objetivo
El objetivo de este ejercicio fue modificar el cliente para que envíe las apuestas en batches, leyendo los datos desde un archivo CSV en lugar de variables de entorno. El servidor debe responder con éxito solo si todas las apuestas del batch fueron procesadas correctamente.

### Cliente (Go)

El cliente ya no lee una apuesta desde variables de entorno sino que abre el archivo `.data/agency-{ID}.csv` inyectado como volumen en `/data/agency-{ID}.csv`. Lee las líneas del archivo en chunks de `batch.maxAmount` (configurable desde `config.yaml`) y por cada chunk envía un mensaje al servidor con todas las apuestas del batch.

El formato del payload se extiende naturalmente del ej5: cada apuesta sigue el formato `agency|first_name|last_name|document|birthdate|number` y las apuestas dentro de un batch se separan con `\n`. El header de 2 bytes con el largo del payload se mantiene igual.

El cliente espera el ACK del servidor antes de enviar el siguiente batch. Si el servidor responde con error, el cliente loguea el fallo y no continúa.

### Servidor (Python)

El servidor recibe el mensaje, separa las apuestas por `\n`, construye la lista de objetos `Bet` y llama a `store_bets(...)` con la lista completa. Si todas las apuestas fueron procesadas correctamente responde `OK` y loguea:

```
action: apuesta_recibida | result: success | cantidad: ${N}
```

En caso de error responde `ERROR` y loguea:

```
action: apuesta_recibida | result: fail | cantidad: 0
```

### Protocolo

No se modificó el formato del header (2 bytes big-endian con el largo del payload). Solo cambió el contenido del payload: en lugar de una apuesta, puede contener N apuestas separadas por `\n`. Esto es posible gracias al diseño extensible del protocolo del ej5.

### `generar-compose.sh`

Se eliminaron las variables de entorno de apuesta (ya no son necesarias) y se agregó el volumen del CSV para cada cliente:

```yaml
- ./.data/agency-{i}.csv:/data/agency-{i}.csv
```

### `config.yaml`

El valor por defecto de `batch.maxAmount` se mantiene en `10`, lo que garantiza que los paquetes no excedan los 8kB considerando el tamaño máximo de cada apuesta serializada.


## Ejercicio 7

### Objetivo
El objetivo de este ejercicio fue agregar la lógica de notificación de fin de envío y consulta de ganadores. El servidor debe esperar a que todas las agencias terminen de enviar sus apuestas antes de realizar el sorteo.

### Protocolo — tipos de mensaje

Se extendió el protocolo agregando un prefijo de tipo al payload. El header de 2 bytes no cambia. Los tipos de mensaje son:

- `B|<apuestas>` — batch de apuestas (igual que ej6, con prefijo `B|`)
- `F` — notificación de fin de envío de una agencia
- `G<agency_id>` — consulta de ganadores de una agencia (ej: `G3`)

El servidor responde:
- A `B` → `OK` o `ERROR`
- A `F` → no responde, solo registra internamente
- A `G` → lista de DNIs ganadores separados por `\n`, string vacío si no hay ganadores, o `NOT_READY` si el sorteo aún no se realizó

### Flujo del cliente

Una vez enviados todos los batches, el cliente:
1. Envía `F` al servidor para notificar que terminó
2. Entra en un loop de polling: envía `G{id}` y si recibe `NOT_READY` espera 500ms y reintenta
3. Al recibir la lista de ganadores loguea: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`

Cada mensaje usa una conexión separada, consistente con el diseño anterior.

### Sincronización en el servidor

El servidor es single-threaded y mantiene un contador `_agencies_done`. Cuando llega un mensaje `F` incrementa el contador. Al llegar a `_total_agencies` realiza el sorteo y setea el flag `_sorteo_done = True`, logueando:

```
action: sorteo | result: success
```

Cuando llega una consulta `G` y el sorteo no está listo, responde `NOT_READY` inmediatamente sin bloquearse, permitiendo que el servidor siga atendiendo otras conexiones. Esto es clave para que el diseño single-threaded funcione correctamente.

### TOTAL_AGENCIES configurable

La cantidad de agencias se configura mediante la variable de entorno `TOTAL_AGENCIES`, que el `generar-compose.sh` setea automáticamente con la cantidad de clientes generados. Esto permite que los tests corran con distintas cantidades de clientes sin modificar el código.

## Ejercicio 8

### Objetivo
El objetivo de este ejercicio fue modificar el servidor para que acepte y procese conexiones en paralelo mediante multithreading.

### Cambios respecto a ej7

El único archivo modificado fue `server/common/server.py`. El cliente no requirió ningún cambio.

**`run()`** — en lugar de manejar cada conexión de forma secuencial, se lanza un thread por conexión:

```python
t = threading.Thread(target=self.__handle_client_connection, args=(client_sock,))
t.daemon = True
t.start()
```

**`_store_lock`** — se agregó un lock para proteger la llamada a `store_bets(...)`, que no es thread-safe.

**`_lock`** — protege el incremento del contador `_agencies_done` para que sea atómico entre threads.

**`_sorteo_event`** — se reemplazó el flag `_sorteo_done` por un `threading.Event`. Los threads que manejan consultas de ganadores hacen `wait()` hasta que el sorteo esté listo, sin necesidad de polling desde el cliente.

### Mecanismos de sincronización

- `_lock` — garantiza que el incremento de `_agencies_done` sea atómico
- `_store_lock` — garantiza acceso exclusivo al archivo de apuestas
- `_sorteo_event` — sincroniza el momento del sorteo con las consultas de ganadores

### Por qué no hay deadlocks

Los dos locks `_lock` y `_store_lock` son independientes y nunca se adquieren juntos en el mismo thread. `_sorteo_event.wait()` no agarra ningún lock mientras espera, por lo que no puede bloquear a ningún otro thread.

### GIL de Python

El servidor usa `threading` en CPython, sujeto al GIL. Sin embargo, las operaciones dominantes son I/O (sockets y archivo), durante las cuales los threads liberan el GIL, permitiendo un paralelismo efectivo para este caso de uso.