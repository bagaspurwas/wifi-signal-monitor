#/bin/bash

#Retrieve current dir
$thisdir = "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

#Preparing
mkdir /etc/wifimonitor
mv $thisdir/config.yaml /etc/wifimonitor/
cp -r ../wifimonitor $GOPATH/src/

#Build golang programwifimonitor
go install wifimonitor
mv $GOPATH/bin/wifimonitor /usr/local/bin/

#Create systemd service instance
cat << 'EOF' > wifimonitor.service
[Unit]
Description= Service to parse wifi signal stregth to InfluxDB

[Service]
Type=
User=root
Group=root
Restart=on-failure
RestartSec=10

ExecStart=/usr/local/bin/wifimonitor

ExecStartPre=/bin/mkdir -p /var/log/wifimonitor
ExecStartPre=/bin/chown syslog:adm /var/log/wifimonitor
ExecStartPre=/bin/chmod 755 /var/log/wifimonitor
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=wifimonitor
tr -d '\n'
[Install]
WantedBy=multi-user.target
EOF

mv wifimonitor.service /etc/systemd/system/

systemctl enable wifimonitor.service
systemctl start wifimonitor.service