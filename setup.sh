#/bin/bash

#Collect parameter and Build a configuration file
POSITIONAL=()
while [[ $# -gt 0]]
do
key="$1"

case $key in
    -d| --db-hostname)
    DBHOSTNAME="$2"
    shift
    shift
    ;;
    -p| --db-port)
    DBPORT="$2"
    shift
    shift
    ;;
    -u| --db-user)
    DBUSER="$2"
    shift
    shift;;
esac
done

#
$thisdir = "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

#Build golang program
export GOPATH = $thisdir
go install wifimonitor

cat << 'EOF' > wifimonitor.service
[Unit]
Description= Service to parse wifi signal stregth to InfluxDB

[Service]
Type=
User=root
Group=root
Restart=on-failure
RestartSec=10

ExecStart=/usr/lib/bin/wifimonitor


ExecStartPre=/bin/mkdir -p /var/log/wifimonitor
ExecStartPre=/bin/chown syslog:adm /var/log/wifimonitor
ExecStartPre=/bin/chmod 755 /var/log/wifimonitor
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=wifimonitor

[Install]
WantedBy=multi-user.target
EOF
