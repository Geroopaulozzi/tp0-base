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

