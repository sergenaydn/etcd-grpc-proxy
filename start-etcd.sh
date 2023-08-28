#!/bin/bash

# Start an etcd instance with the specified configuration.

etcd --name etcd-1 --initial-advertise-peer-urls http://192.168.200.2:2380 \
  --listen-peer-urls http://192.168.200.2:2380 \
  --listen-client-urls http://192.168.200.2:2379 \
  --advertise-client-urls http://192.168.200.2:2379 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster etcd-1=http://192.168.200.2:2380,etcd-2=http://192.168.200.3:2380,etcd-3=http://192.168.200.4:2380 \
  --initial-cluster-state new &

# Sleep for a short duration to allow the etcd instance to start up.

sleep 1

# Start the etcd gRPC proxy.

etcd grpc-proxy start --endpoints=http://192.168.200.2:2380 \
  --listen-addr=0.0.0.0:23790 &

# Sleep again to give time for the gRPC proxy to start.

sleep 1

# Change directory to the application's directory.

cd /app

# Make the 'main' binary executable.

chmod +x /app/main

# Start the main application.

./main

# The following line keeps the script running indefinitely.

while true; do sleep 1000; done
