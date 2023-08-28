FROM  ubuntu:latest

# Install dependencies

RUN apt-get update && apt-get install -y etcd curl
WORKDIR /app
RUN mkdir -p /app
COPY start-etcd.sh /app/start-etcd.sh
RUN chmod +x /app/start-etcd.sh
COPY main .
EXPOSE 2379 2380
ENTRYPOINT ["/app/start-etcd.sh"]
