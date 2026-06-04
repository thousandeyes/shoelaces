/*
  Copyright 2018 ThousandEyes Inc.

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

document.addEventListener('DOMContentLoaded', function () {
    setupNavbarCollapse();
    setVisible(document.getElementById('systems'), false);
    updateHostnames();
    updateEventHistory();

    var target = document.getElementById('target');
    if (target) {
        target.addEventListener('change', scriptSelection);
    }

    window.setTimeout(function () {
        document.querySelectorAll('.alert').forEach(fadeOutAndRemove);
    }, 3000);

    window.setInterval(updateHostnames, 5000);
    window.setInterval(updateEventHistory, 5000);
});

function setupNavbarCollapse() {
    document.querySelectorAll('[data-toggle="collapse"][data-target]').forEach(function (button) {
        var target = document.querySelector(button.getAttribute('data-target'));
        if (!target) {
            return;
        }

        updateCollapseButton(button, target.classList.contains('show'));

        button.addEventListener('click', function () {
            var expanded = !target.classList.contains('show');
            target.classList.toggle('show', expanded);
            updateCollapseButton(button, expanded);
        });
    });
}

function updateCollapseButton(button, expanded) {
    button.setAttribute('aria-expanded', expanded ? 'true' : 'false');
    button.classList.toggle('collapsed', !expanded);
}

function fetchJSON(url) {
    return fetch(url, {
        headers: {
            'Accept': 'application/json',
        },
    }).then(function (response) {
        if (!response.ok) {
            throw new Error('GET ' + url + ' failed: ' + response.status + ' ' + response.statusText);
        }
        return response.json();
    });
}

function logFetchError(error) {
    if (window.console && window.console.error) {
        window.console.error(error);
    }
}

function setVisible(element, visible) {
    if (!element) {
        return;
    }

    if (visible) {
        element.classList.remove('hide');
        element.style.display = '';
    } else {
        element.style.display = 'none';
    }
}

function fadeOutAndRemove(element) {
    element.style.overflow = 'hidden';
    element.style.height = element.scrollHeight + 'px';
    element.style.transition = 'opacity 1s ease, height 1s ease';

    window.requestAnimationFrame(function () {
        element.style.opacity = '0';
        element.style.height = '0';
    });

    window.setTimeout(function () {
        element.remove();
    }, 1000);
}

function updateHostnames() {
    var macs = document.getElementById('mac');
    if (!macs) {
        return;
    }

    fetchJSON('/ajax/servers')
        .then(function (systems) {
            var selected = macs.options[macs.selectedIndex];
            var selection = selected ? selected.textContent : '';

            macs.textContent = '';

            if (systems.length === 0) {
                setVisible(document.getElementById('systems'), false);
                setVisible(document.getElementById('loading'), true);
                return;
            }

            setVisible(document.getElementById('loading'), false);
            setVisible(document.getElementById('systems'), true);

            systems.forEach(function (system) {
                var option = document.createElement('option');
                var systemText = system.Mac + ' - ' + system.IP;

                if (system.Hostname !== '') {
                    systemText += ' - ' + system.Hostname;
                }

                option.className = 'text-primary-custom';
                option.value = system.Mac;
                option.textContent = systemText;
                option.selected = systemText === selection;
                macs.appendChild(option);
            });
        })
        .catch(logFetchError);
}

function scriptSelection() {
    var paramsElems = document.querySelector('.params-container');
    var target = document.querySelector('select[name="target"]');

    if (!paramsElems || !target) {
        return;
    }

    var option = target.options[target.selectedIndex];
    var script = option ? option.dataset.script : '';
    var env = option ? option.dataset.env : '';

    paramsElems.textContent = '';

    if (!script || script.length === 0) {
        return;
    }

    var paramsURL = new URL('/ajax/script/params', window.location.origin);
    paramsURL.searchParams.set('script', script);
    paramsURL.searchParams.set('environment', env);

    fetchJSON(paramsURL.toString())
        .then(function (params) {
            paramsElems.textContent = '';

            params.forEach(function (param) {
                var col = document.createElement('div');
                var input = document.createElement('input');

                col.className = 'col';

                input.type = 'text';
                input.className = 'form-control';
                input.id = param;
                input.name = param;
                input.placeholder = param;
                input.required = true;

                col.appendChild(input);
                paramsElems.appendChild(col);
            });

            var environment = document.createElement('input');
            environment.type = 'hidden';
            environment.name = 'environment';
            environment.value = env;
            paramsElems.appendChild(environment);
        })
        .catch(logFetchError);
}

function updateEventHistory() {
    var eventLogContainer = document.querySelector('.event-log');
    if (!eventLogContainer) {
        return;
    }

    fetchJSON('/ajax/events')
        .then(function (events) {
            if (!events) {
                return;
            }

            eventLogContainer.textContent = '';

            Object.keys(events).forEach(function (mac) {
                eventLogContainer.appendChild(createEventCard(mac, events[mac]));
            });
        })
        .catch(logFetchError);
}

function createEventCard(mac, events) {
    var card = document.createElement('div');
    var header = document.createElement('h5');
    var body = document.createElement('div');
    var list = document.createElement('ul');

    card.className = 'card';
    card.id = mac;

    header.className = 'card-header text-primary-custom';
    header.textContent = eventTitle(mac, events);

    body.className = 'card-body';
    list.className = 'list-group list-group-flush';

    events.forEach(function (event) {
        list.appendChild(createEventItem(event));
    });

    body.appendChild(list);
    card.appendChild(header);
    card.appendChild(body);

    return card;
}

function eventTitle(mac, events) {
    var title = mac;
    var firstEvent = events && events.length > 0 ? events[0] : null;
    var server = firstEvent ? firstEvent.server : null;

    if (!server) {
        return title;
    }

    var host = server.Hostname || server.IP;
    if (host) {
        title += ' (' + host + ')';
    }

    return title;
}

function createEventItem(event) {
    var item = document.createElement('li');
    var date = document.createElement('b');

    item.className = 'list-group-item';
    date.textContent = new Date(event.date).toLocaleString();

    item.appendChild(date);
    item.appendChild(document.createTextNode(': ' + event.message));

    return item;
}
