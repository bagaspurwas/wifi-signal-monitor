#/bin/sh

# Raspberry Pi Deployment ONLY
# This file should be placed on crontab as root
# The function is to make sure that each node run on unique identified name
# and Raspberry Pi connect to desired Wireless SSID

# Retrieve current dir
WD=$(pwd)

DMICECODE="/usr/sbin/dmidecode"

# WiFi 
# Don't forget to edit

DEFAULT_SSID=""
DEFAULT_PSK=""
DEFAULT_PASSPHRASE=""
WPA_SUPPLICANT_FILE="/etc/wpa_supplicant/wpa_supplicant.conf"

#WiFi config location

CONFIG_YAML="/etc/wifimon/config.yaml"

#Setup WiFi
if [ ! -f /etc/network/interfaces.d/wlan0 ]; then
cat << EOT > /etc/network/interfaces.d/wlan0
allow-hotplug wlan0
iface wlan0 inet dhcp
wpa-conf /etc/wpa_supplicant/wpa_supplicant.conf
iface default inet dhcp
EOT
fi

#Create wpa_supplicant configuration file
#if it is not exist or SSID and PASSPHRASE are mismatch

if [ ! -f $WPA_SUPPLICANT_FILE ] || [ -z $(cat $WPA_SUPPLICANT_FILE | grep -E '$DEFAULT_SSID.*$DEFAULT_PASSPHRASE') ]; then
    wpa_passphrase "$DEFAULT_SSID" "$DEFAULT_PASSPHRASE" > $WPA_SUPPLICANT_FILE   
fi


#ENABLE SSH
touch /boot/ssh

# Generate Unique Number
# Use dmidecode to get serial number, ip to get mac address
# and blkid to get UUID and hash three of them using sha256

UNIQUE_ID=$(echo $(sudo dmidecode -t 4 | grep ID | sed 's/.*ID://;s/ //g') \
     $(ip a | grep "ether" | awk -F " " '{print $2, $8}' | head -n 1 | sed 's/://g') \
     $(blkid | grep -oP 'UUID="\K[^"]+' | sha256sum | awk '{print $1}') | sha256sum |
     awk '{print $1}')

echo $UNIQUE_ID

#Modify Configuration File
if [ -f $CONFIG_YAML ] && [ ! -z cat $CONFIG_YAML | grep 'uniqueID: ""' ]; then
	sed -re 's/(uniqueID: ")[^=]/\1'"$UNIQUE_ID"'\"/' -i $CONFIG_YAML
fi

