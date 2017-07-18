#!/bin/bash

set -e # -x

echo "-----> `date`: Upload stemcell"
bosh -n upload-stemcell "https://bosh.io/d/stemcells/bosh-warden-boshlite-ubuntu-trusty-go_agent?v=3421.4" \
  --sha1 e7c440fc20bb4bea302d4bfdc2369367d1a3666e \
  --name bosh-warden-boshlite-ubuntu-trusty-go_agent \
  --version 3421.4

echo "-----> `date`: Delete previous deployment"
bosh -n -d turbulence delete-deployment --force

echo "-----> `date`: Deploy"
( set -e; cd ./..; 
  bosh -n -d turbulence deploy ./manifests/example.yml -o ./manifests/dev.yml \
  -v turbulence_api_ip=10.244.0.34 \
  -v director_ip=192.168.56.6 \
  -v director_client=admin \
  -v director_client_secret=$(bosh int ~/workspace/deployments/vbox/creds.yml --path /admin_password) \
  --var-file director_ssl.ca=<(bosh int ~/workspace/deployments/vbox/creds.yml --path /director_ssl/ca) \
  --vars-store ./tests/creds.yml )

echo "-----> `date`: Deploy dummy"
( set -e; cd ./..; bosh -n -d dummy deploy ./manifests/dummy.yml )

echo "-----> `date`: Kill dummy"
export TURBULENCE_HOST=10.244.0.34
export TURBULENCE_PORT=8080
export TURBULENCE_USERNAME=turbulence
export TURBULENCE_PASSWORD=$(bosh int creds.yml --path /turbulence_api_password)
export TURBULENCE_CA_CERT=$(bosh int creds.yml --path /turbulence_api_cert/ca)
ginkgo -r ./../src/github.com/cppforlife/turbulence-example-test/

echo "-----> `date`: Delete deployments"
bosh -n -d dummy delete-deployment
bosh -n -d turbulence delete-deployment

echo "-----> `date`: Done"
