import pika
import json

rabbitmq_host = 'localhost'  # Replace with your RabbitMQ host if different
connection = pika.BlockingConnection(pika.ConnectionParameters(rabbitmq_host))
channel = connection.channel()

my_queue = 'my_queue'
channel.queue_declare(queue=my_queue)

channel.basic_consume(queue=my_queue, on_message_callback=lambda ch, method, properties, body: print(f"Received message: {json.loads(body)}"), auto_ack=True)
channel.start_consuming()