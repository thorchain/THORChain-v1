#!/bin/bash
# Script to initialize a testnet settings on a server

#Usage: terraform.sh <testnet_name> <testnet_node_number>

#Add thorchaind node number for remote identification
echo "$2" > /etc/thorchaind-nodeid

#Create thorchaind user
useradd -m -s /bin/bash thorchaind

#Reload services to enable the thorchaind service (note that the thorchaind binary is not available yet)
systemctl daemon-reload
systemctl enable thorchaind


