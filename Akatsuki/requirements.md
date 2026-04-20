# Akatsuki
Representa los enemigos del sistema. Van apareciendo a medida que pasa el tiempo. Una gran lista de nombres unicos para er varios a la ves en el sistema de manera dinamica.

## Tareas
### Generar Akatsuki
- Genera de manera periodica los Akatsuki del sistema, con atributos aleatorios, pero aumentan en nivel cada vez que vuelven.

**Implementacion**
- func localizarAkatsuki()
- RabbitMQ, cola 'localizar_akatsuki' para envarlos a ANBU
- El primer mensaje 'CLEAR' permite al sistema el ingreso de Akatsuki, para limpiar las listas existentes en caso de baja del nodo.

### Actualiar Estado
- Mantiana el estado de cada enemido dentro del sistema.

**Implementacion**
- RabbitMQ, cola 'localizar_akatuski' para envar a ANBU
- Solo repite el akatsuki con el nuevo estado.
- Usado por el sistema de combate

### Entregar Lista de Akatsuki a ANBU
Entrega la lista de akatuski a ANBU

**Implementacion**
- func localizarAkatsuki()
- A medida que se genera un nuevo Akatsuki, es inmediatamente enviado a la cola 'localizar_akatsuki'

### Combate
Un Equipo Ninja puede inicializar un combate con un Akatsuki.

**Implementacion**
- func IniciarCombate()
- Implementado usando gRPC para facil sincronizacion

## Maquina Virtual
Se le asigna a este modulo a la siguiente máquina:
- nombre    : dist025
- ip enp1s0 : 10.35.168.35/24    