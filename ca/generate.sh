#!/bin/bash

set -e

certs=`dirname $0`/certs

rm -rf $certs && mkdir -p $certs

cd $certs

echo "Generating CA..."
openssl genrsa -out rootCA.key 2048
yes "" | openssl req -x509 -new -nodes -key rootCA.key \
	-out rootCA.pem -days 99999

function generateCert {
	name=$1
	ip=$2

	cat >openssl-exts.conf <<-EOL
extensions = san
[san]
subjectAltName = IP:${ip}
EOL

	echo "Generating certificate signing request for ${ip}..."
	# golang requires to have SAN for the IP
	openssl req -new -nodes -newkey rsa:2048 \
		-out ${name}.csr -keyout ${name}.key \
		-subj "/C=US/O=BOSH/CN=${ip}"

	echo "Generating certificate ${ip}..."
	openssl x509 -req -in ${name}.csr \
		-CA rootCA.pem -CAkey rootCA.key -CAcreateserial \
		-out ${name}.crt -days 99999 \
		-extfile ./openssl-exts.conf

	echo "Deleting certificate signing request and config..."
	rm ${name}.csr
	rm ./openssl-exts.conf
}

generateCert director 192.168.50.4
generateCert api 10.244.8.2

echo "Finished..."
ls -la .
