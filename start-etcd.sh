#!/bin/bash

# Start etcd proxy
etcd --name etcd1 --initial-advertise-peer-urls http://127.0.0.1:23801 \
  --listen-peer-urls http://127.0.0.1:23801 \
  --listen-client-urls http://127.0.0.1:23791 \
  --advertise-client-urls http://127.0.0.1:23791 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster etcd1=http://127.0.0.1:23801 \
  --initial-cluster-state new &

# Sleep for a few seconds to give the previous command some time to start
sleep 5

# Start etcd grpc-proxy
etcd grpc-proxy start --endpoints=http://127.0.0.1:23791 \
  --listen-addr=0.0.0.0:23790 \
 