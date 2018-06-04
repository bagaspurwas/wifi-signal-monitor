#/bin/bash

#Retrieve current dir
$WD = "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

DMICECODE="/usr/sbin/dmidecode"

#WIFI PARAM
DEFAULT_SSID="test111"
DEFAULT_PSK=""
DEFAULT_PASSPHRASE="test2312324213"
WPA_SUPPLICANT_FILE="/tmp/ccd"

#WiFi config location
CONFIG_YAML="/etc/wifimonitor/config.yaml"

#Setup WiFi

#Create or edit wpa_supplicant configuration file
if [ -z "$(cat $WPA_SUPPLICANT_FILE | grep network=)" ]; then
    wpa_passphrase $DEFAULT_SSID $DEFAULT_PASSPHRASE >> $WPA_SUPPLICANT_FILE
else
    sed 's/\(ssid="\)[^"]*/\1'"$DEFAULT_SSID/"'' $WPA_SUPPLICANT_FILE 
    DEFAULT_PSK=$(wpa_passphrase $DEFAULT_SSID $DEFAULT_PASSPHRASE | grep -e psk=[a-z0-a] | sed 's/^.*=//' )
    echo $DEFAULT_PSK
    sed 's/\(psk=\)[^+]*/\1'"$DEFAULT_PSK/"'' $WPA_SUPPLICANT_FILE 
fi

#Configure WiFi
#wpa_cli -i wlan0 reconfigure

#ENABLE SSH
touch /boot/ssh

# Generate Unique Number
# Use dmidecode to get serial number, ip to get mac address
# and blkid to get UUID and hash three of them using sha256

UNIQUE_ID=$(echo $(sudo dmidecode -t 4 | grep ID | sed 's/.*ID://;s/ //g') \
     $(ip a | grep "ether" | awk -F " " '{print $2, $8}' | head -n 1 | sed 's/://g') \
     $(blkid | grep -oP 'UUID="\K[^"]+' | sha256sum | awk '{print $1}') | sha256sum |
     awk '{print $1}')

#Modify Configuration File

sed -re 's/(uniqueID: ")[^=]/\1'"$UNIQUE_ID"'\"/' $CONFIG_YAML


