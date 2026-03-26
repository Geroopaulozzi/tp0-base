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