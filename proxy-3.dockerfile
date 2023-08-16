FROM ubuntu:latest

RUN apt-get update && apt-get install -y etcd curl

COPY start-etcd-3.sh /start-etcd-3.sh

RUN chmod +x /start-etcd-3.sh

EXPOSE 23793 23803 2380 2379 23790 23791 23792 23793

ENTRYPOINT ["/start-etcd-3.sh"]
