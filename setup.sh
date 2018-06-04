#/bin/bash

#Retrieve current dir
$pwd = "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

#Install Go if it is not installed already
if [ -z "which go 2> /dev/null"]; then
    curl -LO https://github.com/hypriot/golang-armbuilds/releases/download/v1.7.4/go1.7.4.linux-arm64.tar.gz
    tar -xvzf go1.7.4.linux-arm64.tar.gz -C /usr/local
    export PATH=/usr/local/go/bin:$PATH
    echo PATH="/usr/local/go/bin:$PATH" >> ~/.bashrc

#Preparing
mkdir /etc/wifimonitor
mkdir $GOPATH/src/wifimonitor
mv $pwd/config.yaml /etc/wifimonitor/
cp $pwd/main.go $GOPATH/src/wifimonitor

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

# Move service to /etc/systemd/system so that
# it is discovered as systemd service
mv wifimonitor.service /etc/systemd/system/

systemctl start wifimonitor.service
systemctl enable wifimonitor.service