# SD Laboratorio 1
 Laboratorio 1, Sistemas Distribuido

## Intengrantes:
- Cristobal Espinoza Cáceres (202273507-4)
- Benjamín Ponce Carrera     (202173615-8)
- Álvaro Rojas Valenuela     (202273502-3)

### Librerias Go
  "bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
  "strconv"
  "math/rand"
	"sync"
	"time"

  "github.com/rabbitmq/amqp091-go"
  "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

### Instalación
Cada nodo es contenido en su propio Container Docker. Una ves con Docker Engine + Compose preparado, se disponen los siguientes comandos en una terminal alocado en la raiz del proyecto:

**Protocol Buffer**
No debe ser necesario compilar los protoc, dado que vienen compilado por defecto, en caso de necesitarlo se dispone de la siguiente linea:

- make build-protoc

**Akatuski**
Construir y ejecutar Container en VM-025
- make docker-akatsuki

Detener Container
- make stop-docker-akatsuki

**Hokage**
Construir y ejecutar Container en VM-028
- make docker-hokage

Detener Container
- make stop-docker-hokage

**Anbu**
Construir y ejecutar Container en VM-026
- make docker-anbu

Detener Container
- make stop-docker-anbu

**Equipos Ninja**
Construir y ejecutar Container en VM-027
- make docker-equiposninja

Detener Container
- make stop-docker-equiposninja
