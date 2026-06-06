#!/usr/bin/env node

/*
  Copyright 2018-2026 ThousandEyes Inc.

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
*/

'use strict';

const fs = require('fs');
const http = require('http');
const os = require('os');
const path = require('path');
const { spawn } = require('child_process');

const apiURL = process.argv[2] || 'http://localhost:18888';
const chromiumBin = process.env.CHROMIUM_BIN || 'chromium';
const smokeMac = '02:00:00:00:00:42';
const smokeMacPath = smokeMac.replace(/:/g, '-');

const [nodeMajor, nodeMinor] = process.versions.node.split('.').map(Number);
if ((nodeMajor < 22 || (nodeMajor === 22 && nodeMinor < 4)) || typeof WebSocket !== 'function') {
    throw new Error('Node >= 22.4.0 with global WebSocket is required for the frontend smoke test');
}

class CDPClient {
    constructor(webSocketURL) {
        this.nextID = 1;
        this.pending = new Map();
        this.handlers = new Map();
        this.ws = new WebSocket(webSocketURL);

        this.opened = new Promise((resolve, reject) => {
            this.ws.addEventListener('open', resolve, { once: true });
            this.ws.addEventListener('error', reject, { once: true });
        });

        this.ws.addEventListener('message', (event) => {
            const message = JSON.parse(event.data);
            if (message.id && this.pending.has(message.id)) {
                const pending = this.pending.get(message.id);
                this.pending.delete(message.id);
                if (message.error) {
                    pending.reject(new Error(message.error.message));
                } else {
                    pending.resolve(message.result || {});
                }
                return;
            }

            const handlers = this.handlers.get(message.method) || [];
            handlers.forEach((handler) => handler(message));
        });
    }

    async send(method, params, sessionID) {
        await this.opened;

        const id = this.nextID++;
        const message = {
            id: id,
            method: method,
            params: params || {},
        };

        if (sessionID) {
            message.sessionId = sessionID;
        }

        return new Promise((resolve, reject) => {
            this.pending.set(id, { resolve: resolve, reject: reject });
            this.ws.send(JSON.stringify(message));
        });
    }

    on(method, handler) {
        if (!this.handlers.has(method)) {
            this.handlers.set(method, []);
        }
        this.handlers.get(method).push(handler);
    }

    close() {
        this.ws.close();
    }
}

function requestURL(url) {
    return new Promise((resolve, reject) => {
        http.get(url, (response) => {
            let body = '';
            response.setEncoding('utf8');
            response.on('data', (chunk) => {
                body += chunk;
            });
            response.on('end', () => {
                if (response.statusCode < 200 || response.statusCode >= 300) {
                    reject(new Error('GET ' + url + ' failed with ' + response.statusCode));
                    return;
                }
                resolve(body);
            });
        }).on('error', reject);
    });
}

async function requestJSON(url) {
    return JSON.parse(await requestURL(url));
}

async function waitFor(description, check, timeoutMS = 5000) {
    const deadline = Date.now() + timeoutMS;

    while (Date.now() < deadline) {
        if (await check()) {
            return;
        }
        await new Promise((resolve) => setTimeout(resolve, 100));
    }

    throw new Error('Timed out waiting for ' + description);
}

async function waitForDevToolsURL(chromium) {
    let stderr = '';

    return new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
            reject(new Error('Timed out waiting for Chromium DevTools URL. stderr: ' + stderr));
        }, 10000);

        chromium.stderr.on('data', (chunk) => {
            stderr += chunk.toString();
            const match = stderr.match(/DevTools listening on (ws:\/\/[^\s]+)/);
            if (match) {
                clearTimeout(timeout);
                resolve(match[1]);
            }
        });

        chromium.once('exit', (code) => {
            clearTimeout(timeout);
            reject(new Error('Chromium exited before DevTools was ready with code ' + code + '. stderr: ' + stderr));
        });
    });
}

function stopProcess(child) {
    return new Promise((resolve) => {
        if (child.exitCode !== null || child.signalCode !== null) {
            resolve();
            return;
        }

        const timeout = setTimeout(() => {
            child.kill('SIGKILL');
        }, 5000);

        child.once('exit', () => {
            clearTimeout(timeout);
            resolve();
        });

        child.kill('SIGTERM');
    });
}

function formatExceptionDetails(details) {
    const parts = [details.text || 'Unhandled browser exception'];

    if (details.exception) {
        parts.push(details.exception.description || details.exception.value || '');
    }

    if (details.stackTrace && details.stackTrace.callFrames) {
        details.stackTrace.callFrames.forEach((frame) => {
            parts.push(frame.url + ':' + frame.lineNumber + ':' + frame.columnNumber);
        });
    }

    return parts.filter(Boolean).join(' ');
}

async function evaluate(cdp, sessionID, expression) {
    const result = await cdp.send('Runtime.evaluate', {
        expression: expression,
        awaitPromise: true,
        returnByValue: true,
    }, sessionID);

    if (result.exceptionDetails) {
        const exception = result.exceptionDetails.exception;
        throw new Error(
            (exception && (exception.description || exception.value)) ||
            result.exceptionDetails.text ||
            'Runtime evaluation failed'
        );
    }

    return result.result.value;
}

async function waitForPage(cdp, sessionID, expression, description, timeoutMS = 5000) {
    await waitFor(description, async () => {
        return Boolean(await evaluate(cdp, sessionID, expression));
    }, timeoutMS);
}

async function run() {
    const userDataDir = fs.mkdtempSync(path.join(os.tmpdir(), 'shoelaces-chromium-'));
    const chromium = spawn(chromiumBin, [
        '--headless=new',
        '--disable-gpu',
        '--disable-dev-shm-usage',
        '--disable-breakpad',
        '--disable-component-extensions-with-background-pages',
        '--disable-crash-reporter',
        '--disable-crashpad',
        '--disable-extensions',
        '--no-first-run',
        '--no-default-browser-check',
        '--no-crash-upload',
        '--no-sandbox',
        '--remote-debugging-port=0',
        '--user-data-dir=' + userDataDir,
        'about:blank',
    ]);

    let cdp;

    try {
        const browserWS = await waitForDevToolsURL(chromium);
        cdp = new CDPClient(browserWS);
        await cdp.opened;

        const target = await cdp.send('Target.createTarget', { url: 'about:blank' });
        const attached = await cdp.send('Target.attachToTarget', {
            targetId: target.targetId,
            flatten: true,
        });
        const sessionID = attached.sessionId;
        const pageErrors = [];

        cdp.on('Runtime.exceptionThrown', (message) => {
            if (message.sessionId === sessionID) {
                pageErrors.push(formatExceptionDetails(message.params.exceptionDetails));
            }
        });
        cdp.on('Runtime.consoleAPICalled', (message) => {
            if (message.sessionId === sessionID && message.params.type === 'error') {
                pageErrors.push(message.params.args.map((arg) => arg.value || arg.description).join(' '));
            }
        });

        await cdp.send('Runtime.enable', {}, sessionID);
        await cdp.send('Page.enable', {}, sessionID);
        await cdp.send('Emulation.setDeviceMetricsOverride', {
            width: 390,
            height: 844,
            deviceScaleFactor: 1,
            mobile: true,
        }, sessionID);

        await requestURL(apiURL + '/poll/1/' + smokeMacPath + '?host=frontend-smoke');
        await waitFor('seeded unknown server to appear in ajax server list', async () => {
            const servers = await requestJSON(apiURL + '/ajax/servers');
            return servers.some((server) => server.Mac === smokeMac);
        }, 10000);

        await cdp.send('Page.navigate', { url: apiURL + '/' }, sessionID);

        await waitForPage(cdp, sessionID, 'document.readyState === "complete"', 'home page load');
        const collapsedNavbarState = await evaluate(cdp, sessionID, `(() => {
            const button = document.querySelector('.navbar-toggler');
            const menu = document.getElementById('navbarsExample10');
            return {
                buttonDisplay: button ? getComputedStyle(button).display : null,
                menuDisplay: menu ? getComputedStyle(menu).display : null,
                menuOpen: menu ? menu.classList.contains('show') : null,
                expanded: button ? button.getAttribute('aria-expanded') : null,
            };
        })()`);

        if (collapsedNavbarState.buttonDisplay === 'none') {
            throw new Error('Navbar toggler is not visible at mobile width');
        }
        if (collapsedNavbarState.menuDisplay !== 'none' || collapsedNavbarState.menuOpen) {
            throw new Error('Navbar menu is not collapsed before toggling');
        }
        if (collapsedNavbarState.expanded !== 'false') {
            throw new Error('Navbar toggler did not start collapsed');
        }

        const expandedNavbarState = await evaluate(cdp, sessionID, `(() => {
            const button = document.querySelector('.navbar-toggler');
            const menu = document.getElementById('navbarsExample10');
            button.click();
            return {
                menuDisplay: getComputedStyle(menu).display,
                menuOpen: menu.classList.contains('show'),
                expanded: button.getAttribute('aria-expanded'),
            };
        })()`);

        if (expandedNavbarState.menuDisplay === 'none' || !expandedNavbarState.menuOpen) {
            throw new Error('Navbar menu did not open after toggling');
        }
        if (expandedNavbarState.expanded !== 'true') {
            throw new Error('Navbar toggler did not report expanded state');
        }

        const recollapsedNavbarState = await evaluate(cdp, sessionID, `(() => {
            const button = document.querySelector('.navbar-toggler');
            const menu = document.getElementById('navbarsExample10');
            button.click();
            return {
                menuDisplay: getComputedStyle(menu).display,
                menuOpen: menu.classList.contains('show'),
                expanded: button.getAttribute('aria-expanded'),
            };
        })()`);

        if (recollapsedNavbarState.menuDisplay !== 'none' || recollapsedNavbarState.menuOpen) {
            throw new Error('Navbar menu did not collapse after second toggle');
        }
        if (recollapsedNavbarState.expanded !== 'false') {
            throw new Error('Navbar toggler did not report collapsed state');
        }

        await waitForPage(
            cdp,
            sessionID,
            'Boolean(document.querySelector("#mac option[value=\\"' + smokeMac + '\\"]"))',
            'unknown server option',
            10000
        );

        const homeState = await evaluate(cdp, sessionID, `(() => {
            const systems = document.getElementById('systems');
            const loading = document.getElementById('loading');
            return {
                systemsDisplay: systems ? getComputedStyle(systems).display : null,
                loadingDisplay: loading ? getComputedStyle(loading).display : null,
                macText: document.querySelector('#mac option[value="${smokeMac}"]')?.textContent || '',
            };
        })()`);

        if (homeState.systemsDisplay === 'none') {
            throw new Error('Server form did not become visible');
        }
        if (homeState.loadingDisplay !== 'none') {
            throw new Error('Loading panel did not become hidden');
        }
        if (!homeState.macText.includes('frontend-smoke')) {
            throw new Error('Unknown server option did not include hostname: ' + homeState.macText);
        }

        await evaluate(cdp, sessionID, `(() => {
            const mac = document.getElementById('mac');
            const target = document.getElementById('target');
            mac.value = '${smokeMac}';
            target.value = 'flatcar.ipxe';
            target.dispatchEvent(new Event('change', { bubbles: true }));
            return true;
        })()`);

        await waitForPage(
            cdp,
            sessionID,
            'document.querySelectorAll(".params-container input[type=text]").length >= 2',
            'template parameter inputs'
        );

        const paramNames = await evaluate(cdp, sessionID, `Array.from(
            document.querySelectorAll('.params-container input[type=text]'),
            (input) => input.name
        ).sort()`);

        if (JSON.stringify(paramNames) !== JSON.stringify(['cloudconfig', 'version'])) {
            throw new Error('Unexpected parameter inputs: ' + JSON.stringify(paramNames));
        }

        const submitState = await evaluate(cdp, sessionID, `(async () => {
            const form = document.getElementById('systems');
            document.querySelector('input[name="cloudconfig"]').value = 'virtual';
            document.querySelector('input[name="version"]').value = '777.0';
            const response = await fetch(form.action, {
                method: form.method,
                body: new URLSearchParams(new FormData(form)),
            });
            return {
                ok: response.ok,
                status: response.status,
            };
        })()`);

        if (!submitState.ok) {
            throw new Error('Manual selection form POST failed with ' + submitState.status);
        }

        await waitFor('manual selection to leave ajax server list', async () => {
            const servers = await requestJSON(apiURL + '/ajax/servers');
            return !servers.some((server) => server.Mac === smokeMac);
        });

        await cdp.send('Page.navigate', { url: apiURL + '/events' }, sessionID);
        await waitForPage(cdp, sessionID, 'document.readyState === "complete"', 'events page load');
        await waitForPage(
            cdp,
            sessionID,
            'document.querySelector(".event-log .card") && document.body.textContent.includes("' + smokeMac + '")',
            'event history card'
        );

        const eventText = await evaluate(cdp, sessionID, 'document.querySelector(".event-log").textContent');
        if (!eventText.includes('A user selected flatcar.ipxe')) {
            throw new Error('Event history did not render user selection: ' + eventText);
        }

        if (pageErrors.length > 0) {
            throw new Error('Browser console errors: ' + pageErrors.join('; '));
        }
    } finally {
        if (cdp) {
            cdp.close();
        }
        await stopProcess(chromium);
        fs.rmSync(userDataDir, { recursive: true, force: true });
    }
}

run().catch((error) => {
    console.error(error.stack || error.message);
    process.exit(1);
});
