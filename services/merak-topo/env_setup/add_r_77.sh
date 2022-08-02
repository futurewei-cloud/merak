ssh root@10.213.43.195 "route add -net 10.200.0.0/16 gw 10.244.6.131"
echo "10.213.43.195 route added"

ssh root@10.213.43.194 "route add -net 10.200.0.0/16 gw 10.244.2.79"
echo "10.213.43.194 route added"

ssh root@10.213.43.78 "route add -net 10.200.0.0/16 gw 10.244.1.140"
echo "10.213.43.78 route added"

ssh root@10.213.43.197 "route add -net 10.200.0.0/16 gw 10.244.3.106"
echo "10.213.43.197 route added"

ssh root@10.213.43.196 "route add -net 10.200.0.0/16 gw 10.244.5.100"
echo "10.213.43.196 route added"

ssh root@10.213.43.193 "route add -net 10.200.0.0/16 gw 10.244.4.46"
echo "10.213.43.193 route added"
