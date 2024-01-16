#!/bin/bash
# set/source env vars first.
# $1 = domain name
virsh emulatorpin $1 | virsh-json