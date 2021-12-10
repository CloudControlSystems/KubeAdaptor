#!/bin/bash
pscp -l root -h node_ip.txt /etc/etcd/ssl/etcd.pem /etc/etcd/ssl/etcd-key.pem /etc/etcd/ssl
