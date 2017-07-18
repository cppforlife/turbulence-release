#!/bin/bash

body='
{
	"Schedule": "@every 1m",

	"Incident": {
		"Tasks": [{
			"Type": "Stress",
			"Timeout": "30s",

			"NumCPUWorkers": 1
		}],

		"Selector": {
			"Deployment": {
				"Name": "dummy"
			},
			"Group": {
				"Name": "dummy_*"
			},
			"ID": {
				"Limit": "0%-50%"
			}
		}
	}
}
'

echo $body | curl -vvv -k -X POST https://turbulence:${TURBULENCE_PASSWORD}@10.244.0.34:8080/api/v1/scheduled_incidents -H 'Accept: application/json' -d @-

echo
