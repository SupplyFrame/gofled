#!/usr/bin/env bash

cd $(dirname $0)

if [[ -f gofled.service ]]; then
	echo "Enabling Service..."
	systemctl enable $(pwd)/gofled.service || exit -1

	echo "Starting Service..."
	systemctl start gofled.service
	systemctl status gofled.service
else
	echo "Could not find gofled.service."
	exit -1
fi
