# Anbu
Unidad de iteligencia, monitorea la activdad de *Akatsuki*, detecta la presencia de enemigos y reporta al *Hokage*. Actúa como un sistema de vigilancia que recopila información sobre los enemigos y genera eventos quie alimentan el resto del sistema.

## Tareas
### Localizar Akatsuki
Recibe información sobre los miembros activos de *Akatsuki* desde la entidad *Akatsulki*. Es decir, almacena un registro actualizado de los enemigos existentes en el sistema.

- De manera **ASYNC**, *Akatsuki* envía la lista de enemigos disponibles al sistema de inteligencia *Anbu*.
- Un akatsuki posee:
    - Id    : Int
    - Nombre: String
    - Ataque: Int
    - Vida:   Int
    - Estado: Enum Estado (Localizado, En Combate, Capturado)

### Entregar Lista de Akatsuki a Hokage
Publica reportes de iteligencia hacia el *Hokage*, informando sobre enemigos detectados, su estado o cualquier cambio relevante en la información disponible.

### Entregar Estado de Akatsuki que entren en Combate
Cuando un miembro de *Akatsuki* entra en combate con un *Equipo ninja* y al *Hokage*. Este evento es publicado para infromar al sistema que el enemigo ya se encuentra ocupado y evitar que otros equipos intenten enfrentarlo simultáneamente.

## RabbitMQ
La maquina de Anbu contendrá el servicio de RabbitMQ para todo el sistema.
Puertos:
- "5672:5672"
- "15672:15672"

## Maquina Virtual
Se le asigna a este modulo a la siguiente máquina:
- nombre    : dist026
- ip enp1s0 : 10.35.168.36/24    