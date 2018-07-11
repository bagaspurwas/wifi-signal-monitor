from flask import Flask, render_template, redirect
from flask import request as r
import yaml
import regex as re
from collections import OrderedDict

app = Flask(__name__)
defaultPath = "/etc/wifimon/config.yaml"

# Config class definition

class Configuration:
    configdict = None
    def __init__(self, path=defaultPath):
        self.path = path

    def loadConfig(self):
        try:
            with open(self.path, 'r') as cf:
                cfdict = yaml.load(cf)
                self.configdict = cfdict
                cf.close()
            return 1
        except:
            print("Load config failed!")
            return 0

    def updateConfig(self, config=None):
        try:
            with open(self.path, 'w') as yaml_config:
                yaml.dump(config, yaml_config, default_flow_style=False)
                cf.close()
            return 1
        except:
            return 0

# FLASK APPS

@app.route("/")
def redirectto():
    return redirect("/index.html")

@app.route("/index.html", methods=['get'])
def defaultView():
    if r.method =='GET':
        config = Configuration()
        if not config.loadConfig():
            return "ERROR OPENING FILE"
        else:
            return render_template("index.html", config=config.configdict)

@app.route("/index.html", methods=['post'])
def pushChange():
    config = Configuration()
    if r.method=='POST':
        configDict = OrderedDict()
        configDict['wireless'] = OrderedDict([
            ('SSID', r.form.get('SSID')),
            ('Passphrase', r.form.get('Passphrase'))
        ])
        configDict['probeNode'] = OrderedDict([
            ('uniqueID', r.form.get('uniqueID')),
            ('alias', r.form.get('alias')),
            ('wlanInterface', r.form.get('wlanInterface')),
            ('location', r.form.get('location'))
        ])
        configDict['influxdb'] = OrderedDict([
            ('host', r.form.get('host')),
            ('port', r.form.get('port')),
            ('username', r.form.get('username')),
            ('password', r.form.get('password')),
            ('database', r.form.get('database')),
            ('measurement', r.form.get('measurement')),
            ('retentionPolicy', r.form.get('retentionPolicy'))
        ])

        represent_dict_order = lambda self, data:  self.represent_mapping('tag:yaml.org,2002:map', data.items())
        yaml.add_representer(OrderedDict, represent_dict_order)

        if config.updateConfig(configDict):
            return "Config file updated"
        else:
            return "Something going wrong. Update not successful"

if __name__ == '__main__':
   app.run(port=5000, debug = True)
