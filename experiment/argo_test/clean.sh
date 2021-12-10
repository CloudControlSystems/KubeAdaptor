kubectl delete -f ./deploy/rbac-argo.yaml
kubectl delete -f ./deploy/resourceUsage.yaml
cd ./deploy
rm Montage.yaml
#rm Epigenomics.yaml
#rm LIGO.yaml
#rm CyberShake.yaml
rm resourceUsage.yaml
rm rbac-argo.yaml
cd .. && cd ./delpods
./delArgoPods.sh
