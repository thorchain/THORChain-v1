#!/bin/bash
# Script to initialize a testnet settings on a server

#Usage: terraform.sh <testnet_name> <testnet_node_number>

#Add thorchaind node number for remote identification
echo "$2" > /etc/thorchaind-nodeid

#Create thorchaind user
useradd -m -s /bin/bash thorchaind
#cp -r /root/.ssh /home/thorchaind/.ssh
#chown -R thorchaind.thorchaind /home/thorchaind/.ssh
#chmod -R 700 /home/thorchaind/.ssh

#Reload services to enable the thorchaind service (note that the thorchaind binary is not available yet)
systemctl daemon-reload
systemctl enable thorchaind


