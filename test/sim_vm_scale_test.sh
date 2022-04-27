#!/bin/bash

x=1
while [ $x -le 10000 ]
do
  ip netns add test_$x
  ip link add veth_$x type veth peer name peer_$x
  ip link set veth_$x netns test_$x
  echo $x
  x=$(( $x + 1 ))
done
