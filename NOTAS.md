## RabbitMQ
Dowload and start rabbitmq using docker, see the local port is used
                                       ¬ local ¬
docker run -it --rm --name rabbitmq -p 5672:5672 -p 15672:15672 rabbitmq:4-management

Import rabbitmq in GO
amqp "github.com/rabbitmq/amqp091-go"

### Usage:

**Conection to AMQP**
con, err := amqp.Dial("amqp ip")

**Channel opening**
channel, err := conn.Channel()

**Queue Declaration**
queue, err := channel.QueueDeclare(
		"tareas", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // args
	)

##### Producer Side

**Publish mensage**
channel.Publish(
    "",         // exchange
    queue.Name, // routing key
    false,      // mandatory
    false,      // immediate
    amqp.Publishing{
        ContentType: "text/plain",
        Body:        []byte(body),
    },
)

#### Consumer Side
msgs, err := channel.Consume(
    queue.Name, // queue
    "",         // consumer
    true,       // auto-ack
    false,      // exclusive
    false,      // no-local
    false,      // no-wait
    nil,        // args
)