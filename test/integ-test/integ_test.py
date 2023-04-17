#!/usr/bin/env python3

# Copyright 2018 ThousandEyes Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

""" Test shoelaces """

import os
import signal
import subprocess
import sys
import time
import tempfile
import string
import pytest
import requests
import datetime
import dateutil.parser
from requests.exceptions import RequestException

API_ADDR = 'localhost:18888'
API_URL = "http://{}".format(API_ADDR)
TEST_DIR = os.path.dirname(os.path.abspath(__file__))
BASE_DIR = os.path.dirname(os.path.dirname(TEST_DIR))
FIXTURE_DIR = os.path.join(TEST_DIR, 'expected-results')
STATIC_DIR = os.path.join(BASE_DIR, "web")
SHOELACES_BINARY = os.path.join(BASE_DIR, "shoelaces")


@pytest.fixture(scope="session", autouse=True)
def shoelaces_binary():
    os.chdir(BASE_DIR)
    subprocess.check_call(["go", "build"])
    os.chdir(TEST_DIR)


@pytest.fixture(scope="session", autouse=True)
def config_file(shoelaces_binary):
    """ Create a temporary config file """
    temp_config_tpl = string.Template("bind-addr=$bind_addr\n"
                                      "data-dir=integ-test-configs\n"
                                      "static-dir=$static_dir\n"
                                      "template-extension=.slc\n"
                                      "mappings-file=mappings.yaml\n"
                                      "debug=true\n")
    temp_config = temp_config_tpl.substitute(bind_addr=API_ADDR,
                                             static_dir=STATIC_DIR)

    sys.stderr.write("Using:\n{}".format(temp_config))
    temp_cfg_file = tempfile.NamedTemporaryFile(delete=False)
    temp_cfg_file.write(bytes(temp_config, 'ascii'))
    temp_cfg_file.flush()
    temp_cfg_file_name = temp_cfg_file.name
    temp_cfg_file.close()
    yield temp_cfg_file_name
    os.unlink(temp_cfg_file_name)


@pytest.fixture(scope="session", autouse=True)
def shoelaces_instance(config_file):
    """ Shoelaces test fixture. """
    shoelaces_start_cmd = [SHOELACES_BINARY, "-config", config_file]
    shoelaces = subprocess.Popen(shoelaces_start_cmd, preexec_fn=os.setsid)
    sys.stderr.write("\nStarting Shoelaces...\n")
    yield shoelaces
    sys.stderr.write("\nShutting down Shoelaces...\n")
    os.killpg(os.getpgid(shoelaces.pid), signal.SIGTERM)
    sys.stderr.write("\nDone\n")


def test_shoelaces_startup(shoelaces_instance):
    """ Test API liveness """
    attempts = 0
    while True:
        try:
            req = requests.get('{}/'.format(API_URL))
            req.raise_for_status()
            sys.stderr.write('\n\nApi startup successful.\n')
            break
        except RequestException:
            attempts += 1
            if attempts > 10:
                raise
            sys.stderr.write(".")
            time.sleep(1)


@pytest.mark.parametrize(("path"), [("/"), ("/events"), ("/mappings")])
def test_response_success(shoelaces_instance, path):
    r = requests.get("{}{}".format(API_URL, path))
    r.raise_for_status()


REQUEST_RESPONSE_PAIRS = [("/static/", "static.html"),
                          ("/configs/static/", "configs-static-default.txt"),
                          ("/configs/static/rc.local-bootstrap",
                           "rc.local-bootstrap"),
                          ("/start", "start.txt"),
                          ("/ipxemenu", "ipxemenu.txt")]


@pytest.mark.parametrize(("request_path", "response_file"), REQUEST_RESPONSE_PAIRS)
def test_request_response(shoelaces_instance, request_path, response_file):
    with open(os.path.join(FIXTURE_DIR, response_file)) as response_body:
        assert requests.get(
            API_URL + request_path).text == response_body.read()


def gen_mac_server_pairs():
    generated = []
    for m in range(0x00, 0x100, 0x11):
        o = "{:02x}".format(m)
        generated.append({'IP': '127.0.0.1', 'Mac': "ff:ff:ff:ff:ff:{}".format(o), 'Hostname': 'localhost'})
        yield (o, list(generated))


@pytest.mark.parametrize(("mac_last_octet", "servers"), gen_mac_server_pairs())
def test_servers(shoelaces_instance, mac_last_octet, servers):
    def sort_by_mac(srv):
        return srv['Mac']

    poll_url = "{}/poll/1/ff-ff-ff-ff-ff-{}".format(API_URL, mac_last_octet)
    req = requests.get(poll_url)
    req = requests.get("{}/ajax/servers".format(API_URL))
    assert sorted(req.json(), key=sort_by_mac) == sorted(servers, key=sort_by_mac)


def test_unknown_server(shoelaces_instance):
    poll_url = "{}/poll/1/06-66-de-ad-be-ef".format(API_URL)
    # Request for unknown host will give result in retries/polling
    with open(os.path.join(FIXTURE_DIR, "poll-unknown.txt")) as poll:
        assert requests.get(poll_url).text == poll.read()
    # Setting the config for the new host should succeed.
    requests.post(API_URL + '/update/target',
                  {"target": "coreos.ipxe",
                   "mac": "06:66:de:ad:be:ef",
                   "version": "666.0",
                   "cloudconfig": "virtual"}).raise_for_status()
    # After setting we should be able to get the new config.
    with open(os.path.join(FIXTURE_DIR, "poll-unknown-set-from-ui.txt")) as poll:
        assert requests.get(poll_url).text == poll.read()
    # Once fetched the host is now again "unknown"
    with open(os.path.join(FIXTURE_DIR, "poll-unknown.txt")) as poll:
        assert requests.get(poll_url).text == poll.read()


def test_events(shoelaces_instance):
    url = "{}/ajax/events".format(API_URL)
    req = requests.get(url)
    req.raise_for_status()
    res = req.json()
    # assert mac is in dictionary
    assert '06:66:de:ad:be:ef' in res
    # assert array with one element
    assert isinstance(res['06:66:de:ad:be:ef'], list) and len(res['06:66:de:ad:be:ef']) == 4
    # assert we have a date field
    assert 'date' in res['06:66:de:ad:be:ef'][0]
    # assert our date actually parses
    assert dateutil.parser.parse(res['06:66:de:ad:be:ef'][0]['date'])
    del res['06:66:de:ad:be:ef'][0]['date']
    # compare to the expected result sans the date as it would be different
    assert sorted(res['06:66:de:ad:be:ef'][0]) == sorted({'eventType': '0',
                                                          'message': '0',
                                                          'bootType': 'Manual',
                                                          'server': {'mac':'',
                                                                     'ip': '',
                                                                     'hostname': '06-66-de-ad-be-ef'},
                                                          'params': {'baseURL': 'localhost:18888',
                                                                     'cloudconfig': 'virtual',
                                                                     'hostname': '06-66-de-ad-be-ef',
                                                                     'version': '666.0'},
                                                          'script': 'coreos.ipxe'})


POLL_PAIRS = [(None, "poll.txt"),
              ({"host": "k8s1-3"}, "poll-k8s1-3-stg.txt"),
              ({"host": "k8s1-4"}, "poll-k8s1-4-stg.txt"),
              ({"host": "k8s1-1"}, "poll-k8s1-1.txt"),
              ({"host": "k8s1-2"}, "poll-k8s1-2.txt")]


@pytest.mark.parametrize(("params", "expected"), POLL_PAIRS)
def test_poll(shoelaces_instance, params, expected):
    """ Test Poll handler """
    poll_url = "{}/poll/1/ff-ff-ff-ff-ff-ff".format(API_URL)
    req = requests.get(poll_url, params=params)
    req.raise_for_status()
    with open(os.path.join(FIXTURE_DIR, expected), 'r') as poll:
        assert poll.read() == req.text


TPL_VARS_PAIRS = [("coreos.ipxe", "", ["cloudconfig", "version"]),
                  ("coreos.ipxe", "default", ["cloudconfig", "version"]),
                  ("coreos.ipxe", "production", ["cloudconfig", "version", "hostname"])]


@pytest.mark.parametrize(("script", "env", "vars"), TPL_VARS_PAIRS)
def test_template_variables_list(shoelaces_instance, script, env, vars):
    url = "{}/ajax/script/params".format(API_URL)
    req = requests.get(url, params={"script": script, "environment": env})
    req.raise_for_status()
    assert sorted(req.json()) == sorted(vars)


if __name__ == "__main__":
    pytest.main(args=sys.argv[1:], plugins=None)
