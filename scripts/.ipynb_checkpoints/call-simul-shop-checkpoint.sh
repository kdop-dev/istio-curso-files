#!/bin/bash

# Configurando acesso ao Ingress
export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
export INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].port}')
export SECURE_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="https")].port}')
export TCP_INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="tcp")].port}')

while true
do
## Sem modificar o /etc/hosts
#curl -v --header "Host: simul-shop.com" http://localhost/s
#curl -v -H "Host: www.simul-shop.com" --resolve "www.simul-shop.com:80:127.0.0.2" http://www.simul-shop.com/s
## Resolvendo pelo /etc/hosts
# curl -v http://simul-shop.com/s
curl -v -H "Host: www.simul-shop.com" http://$INGRESS_HOST:$INGRESS_PORT/s
echo
sleep 1
done