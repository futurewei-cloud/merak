#!/bin/bash

service rsyslog restart && /etc/init.d/openvswitch-switch restart && /merak/services/merak-agent/plugins/alcor-control-agent/build/bin/AlcorControlAgent -d -a 10.213.43.224 -p 30014