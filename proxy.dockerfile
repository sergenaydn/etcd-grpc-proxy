FROM ubuntu:latest

RUN apt-get update && apt-get install -y etcd curl

EXPOSE 2379 2380 23791 23801 23790

COPY start-etcd-proxy.sh /start-etcd-proxy.sh

RUN chmod +x /start-etcd-proxy.sh

CMD ["/start-etcd-proxy.sh"]
    