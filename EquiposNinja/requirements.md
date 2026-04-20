# EquposNinja
Interfaz de usuario con el SD. Se inicializa cada isntancia de Equipo Ninja con una nueva instancia de Terminal. Puede combatir a los akatsukis y recibir recompenzas desde el Hokage

## Tareas
### Crear Equipo
Inicializa la instancia de Equipo Ninja al sistema, con atributos dado por el usuario

**Implementacion**
- Para que el equipo ninja pueda recivir las actualizaciones de ANBU, se usa una subcripcion a 'EquiposNinjaBrodcast' usando QueueBind() y una cola sin nombre que es uso personal.


### Mostrar Estadistica
Muestra los atributos del equipo.

### Escuchar Info de los ABU
- Escucha a ANBU en caso de cambios de estados de algun akatsuki o nuevas entradas.

**Implementacion**
- Usa el modelos descrito en Crear Equipo, solo actualiza los estados o nuevas entradas en su lista local en RAM.

### Solicitar Lista de Hokage
- Solicita la lista completa de Akatsuki disponibles en el sistema en caso de que es nuevo en el sistema.

**Implementacion**
- Usa una query gRPC a hokage para obtener la lista, mejora dado por la necesidad de Sync

### Atacar Akatsuki
- Ataca a un akatsuki.

**Implementacion**
- Solicita el ataque mediente un gRPC para esperar el resultado, enviando la informacion de ataque, vida y objetivo.

### Solicitar Recompensa de Hokage
- Solicita la recompensa a Hokage por un akatsuki capturado.

**Implementacion**
- Usa una query gRPC a hokage para obtener la recompensa, mejora dado por la necesidad de Sync


## Maquina Virtual
Se le asigna a este modulo a la siguiente máquina:
- nombre    : dist027
- ip enp1s0 : 10.35.168.37/24    