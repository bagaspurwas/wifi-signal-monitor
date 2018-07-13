from flask import Flask, render_template, redirect, session, request, abort, flash
from flask import request as r
import yaml
import regex as re
import os
from collections import OrderedDict

app = Flask(__name__)
defaultPath = "../config.yaml"

# Serialize yaml to ORderedDict, Python 3.6<
# Replace yaml.dump(stream) to orderedLoad(stream, yaml.SafeLoader)

def orderedLoad(stream, Loader=yaml.Loader, object_pairs_hook=OrderedDict):
    class OrderedLoader(Loader):
        pass
    def construct_mapping(loader, node):
        loader.flatten_mapping(node)
        return object_pairs_hook(loader.construct_pairs(node))
    OrderedLoader.add_constructor(
        yaml.resolver.BaseResolver.DEFAULT_MAPPING_TAG,
        construct_mapping)
    return yaml.load(stream, OrderedLoader)

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
                yaml_config.close()
            return 1
        except:
            return 0

# FLASK APPS

@app.route("/")
def redirectto():
    return redirect("/index.html")

@app.route("/login", methods=['post'])
def login():
    if (request.form['password'] == 'password' and request.form['username'] == 'admin'):
        session['logged_in'] = True
    else:
        flash('Wrong username or password!')
    return redirect("index.html")

@app.route("/logout")
def logout():
    session['logged_in'] = False
    return defaultView()

@app.route("/index.html", methods=['get'])
def defaultView():
    if (r.method =='GET' and session.get('logged_in')):
        config = Configuration()
        if not config.loadConfig():
            return "ERROR OPENING FILE"
        else:
            ordconf = OrderedDict(config.configdict)
            return render_template("index.html", config=ordconf)
    else:
        return render_template("login.html")

@app.route("/index.html", methods=['post'])
def pushChange():
    if (r.method=='POST' and session.get('logged_in')):
        config = Configuration()
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
    else:
        return render_template("login.html")

if __name__ == '__main__':
    app.secret_key = os.urandom(12)
    app.run(port=5000, debug = True)
