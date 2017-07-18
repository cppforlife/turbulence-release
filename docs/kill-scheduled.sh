#!/bin/bash

body='
{
	"Schedule": "@every 1m 30s",

	"Incident": {
		"Tasks": [{
			"Type": "Kill"
		}],

		"Selector": {
			"Deployment": {
				"Name": "dummy"
			},
			"Group": {
				"Name": "dummy_*"
			},
			"ID": {
				"Limit": "0%-20%"
			}
		}
	}
}
'

echo $body | curl -vvv -k -X POST https://turbulence:${TURBULENCE_PASSWORD}@10.244.0.34:8080/api/v1/scheduled_incidents -H 'Accept: application/json' -d @-

echo
