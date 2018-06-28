#!/bin/sh

$GOPATH=$GOPATH:$HOME/go
which=/bin/which
#Create a config file dir and copy
mkdir /etc/wifimon
cp ./config.yaml /etc/wifimon/

#install the program using go
if  [ ! -z `which go 2> /dev/null` ] and [ "$1" != "precompiled" ]; then
  echo "Building program..."
  go build wifi-sign2l-monitor-master -o wifimon
  mv $GOPATH/bin/wifimon /etc
elif [ "$1"=="precompiled" ]; then
  echo "Using precompiled binary..."
  cp bin/arm7/wifimon /usr/local/bin/wifimon
else
    echo "Go not found!"
    exit 1
fi
#create systemd service
cat << 'EOF' > wifimon.service
[Unit]
Description=Service to parse wifi signal strength to InfluxDB

[Service]
Type=idle
User=root
Group=root
Restart=on-failure
RestartSec=10

ExecStart=/usr/local/bin/wifimon

ExecStartPre=/bin/mkdir -p /var/log/wifimon
ExecStartPre=/bin/chown root:root /var/log/wifimon
ExecStartPre=/bin/chmod 755 /var/log/wifimon
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=wifimon

[Install]
WantedBy=multi-user.target
EOF
# move, start, and enable the service
mv wifimon.service /etc/systemd/system/
systemctl start wifimon
systemctl enable wifimons
