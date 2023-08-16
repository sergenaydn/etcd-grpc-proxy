FROM ubuntu:latest

RUN apt-get update && apt-get install -y etcd curl

COPY start-etcd.sh /start-etcd.sh

RUN chmod +x /start-etcd.sh

EXPOSE  23791 23801 2379 2380 23790 23791 23792 23793

ENTRYPOINT ["/start-etcd.sh"]
