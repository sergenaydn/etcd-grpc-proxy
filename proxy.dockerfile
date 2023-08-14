FROM ubuntu:latest

RUN apt-get update && apt-get install -y etcd curl

COPY start-etcd.sh /start-etcd.sh
RUN chmod +x /start-etcd.sh

EXPOSE 2379 2380 23791 23801 23790

CMD ["/start-etcd.sh"]
