#!/bin/bash

etcd --name etcd-2 --initial-advertise-peer-urls http://192.168.200.3:2380 \
  --listen-peer-urls http://192.168.200.3:2380 \
  --listen-client-urls http://192.168.200.3:2379\
  --advertise-client-urls http://192.168.200.3:2379 \
  --initial-cluster-token etcd-cluster-1 \
  --initial-cluster etcd-1=http://192.168.200.2:2380,etcd-2=http://192.168.200.3:2380,etcd-3=http://192.168.200.4:2380 \
  --initial-cluster-state new


sleep 5

# Start etcd grpc-proxy
etcd grpc-proxy start --endpoints=http://192.168.200.3:2380 \
  --listen-addr=0.0.0.0:2379 &
 while true; do sleep 1000; done