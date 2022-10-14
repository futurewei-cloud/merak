for d in services
do
	kubectl delete -f $d
	#sed -i 's/replicas: 9/replicas: 1/g' $d
done
for d in db/ignite
do
        kubectl delete -f $d
	#sed -i 's/replicas: 9/replicas: 1/g' $d
done
