FROM  ubuntu

# Install dependencies

RUN apt-get update && apt-get install -y etcd curl
WORKDIR /app
RUN mkdir -p /app
COPY start-etcd-2.sh /app/start-etcd-2.sh
RUN chmod +x /app/start-etcd-2.sh
COPY main .
EXPOSE 2379 2380
ENTRYPOINT ["/app/start-etcd-2.sh"]
