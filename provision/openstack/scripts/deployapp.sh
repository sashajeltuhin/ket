#!/bin/bash
mkdir /root/.kube
cd /ket
cp -p generated/kubeconfig ~/.kube/config
kubectl create -f /ket/osrm.yaml
kubectl create -f /ket/geo-service.json
kubectl create -f /ket/geo-ingress.yaml