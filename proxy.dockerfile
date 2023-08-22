FROM  ubuntu as etcd

# Install dependencies
RUN apt-get update && apt-get install -y etcd curl wget
COPY start-etcd.sh /app/start-etcd.sh
RUN mkdir -p /app
# Change permissions for script
RUN chmod +x /app/start-etcd.sh


FROM ubuntu as app
RUN apt-get update && apt-get install -y etcd curl wget

# Create app directory
RUN mkdir -p app/ginapp

# Download and install Go
RUN wget https://dl.google.com/go/go1.21.0.linux-amd64.tar.gz -O app/ginapp/go1.21.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf app/ginapp/go1.21.0.linux-amd64.tar.gz

# Set environment variables
ENV PATH "$PATH:/usr/local/go/bin/"
ENV GOROOT=/usr/local/go/
ENV GOPATH=/go

RUN mkdir -p /app
# Copy necessary files
WORKDIR /app
COPY go.mod go.sum ./
COPY . .
COPY start-etcd.sh /app/start-etcd.sh

# Change permissions for script
COPY --from=etcd . /app/ginapp
# Expose necessary ports
EXPOSE 2379 2380
RUN chmod +x /app/start-etcd.sh

# Set the entry point
ENTRYPOINT ["/app/start-etcd.sh"]
