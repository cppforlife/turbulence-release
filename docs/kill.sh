#!/bin/bash

set -e

body='
{
	"Tasks": [{
		"Type": "kill"
	}],

	"Selector": {
		"Deployment": {
			"Name": "dummy"
		},
		"Group": {
			"Name": "dummy_z1"
		},
		"ID": {
			"Limit": "10%-60%"
		}
	}
}
'

echo $body | curl -vvv -k -X POST https://turbulence:${TURBULENCE_PASSWORD}@10.244.0.34:8080/api/v1/incidents -H 'Accept: application/json' -d @-

echo
