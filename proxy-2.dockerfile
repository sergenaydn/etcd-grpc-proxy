FROM ubuntu:latest

RUN apt-get update && apt-get install -y etcd curl

COPY start-etcd-2.sh /start-etcd-2.sh

RUN chmod +x /start-etcd-2.sh

EXPOSE 23792 23802 2380 2379 23790 23791 23792 23793

ENTRYPOINT ["/start-etcd-2.sh"]
