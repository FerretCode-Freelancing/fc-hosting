from kubemq.queue.message_queue import MessageQueue
from kubemq.queue.message import Message
from kubernetes import client, config
from random import randrange

import socket
import os

def connect():
  dns = "kubemq-cluster-grpc.kubemq.svc.cluster.local" 

  ip = socket.gethostbyname_ex(dns)[2][0] 

  channel = "status-updates"

  queue = MessageQueue(
    channel,
    randrange(1, 25000), 
    f"{ip}:50000" 
  )

  return queue

queue = connect()
config.load_incluster_config()
namespace = os.getenv('POD_NAMESPACE')

while True: 
  v1 = client.CoreV1Api()

  pods = v1.list_namespaced_pod(namespace)

  for pod in pods.items:
    print(pod.metadata.name)
