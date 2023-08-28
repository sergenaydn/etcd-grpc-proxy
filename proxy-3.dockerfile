FROM  ubuntu:latest


RUN apt-get update && apt-get install -y etcd curl
WORKDIR /app
RUN mkdir -p /app
COPY start-etcd-3.sh /app/start-etcd-3.sh
RUN chmod +x /app/start-etcd-3.sh
COPY main .
EXPOSE 2379 2380
ENTRYPOINT ["/app/start-etcd-3.sh"]
