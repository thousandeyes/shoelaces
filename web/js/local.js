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

$(document).ready(function () {
    $('#systems').hide()
    updateHostnames();
    updateEventHistory();
    $('#target').on('change', scriptSelection);

    window.setTimeout(function () {
        $('.alert').fadeTo(1000, 0).slideUp(1000, function () {
            $(this).remove();
        });
    }, 3000);
});

window.setInterval(updateHostnames, 5000);
window.setInterval(updateEventHistory, 5000);

function updateHostnames() {
    $.getJSON('/ajax/servers', function (systems) {
        var macs = $('#mac');
        var selection = $('select[name="mac"]').find('option:selected').text();
        macs.empty();

        if (systems.length == 0) {
            $('#systems').fadeOut(500);
            $('#loading').fadeIn(500);
        } else {
            $('#loading').hide(500);
            $('#systems').removeClass('hide');
            $('#systems').fadeIn(500);

            $.each(systems, function () {
                var system_str = this.Mac + ' - ' + this.IP;
                if (this.Hostname != '') {
                    system_str += ' - ' + this.Hostname;
                }
                macs.append('<option class="text-primary-custom" value="' + this.Mac + '">' + system_str  +  '</option>');
            });

            $('#mac option').filter(function () {
                //may want to use $.trim in here
                return $(this).text() == selection;
            }).prop('selected', true);
        }
    });
}

function scriptSelection() {
    var paramsElems = $('.params-container');
    var option = $('select[name="target"]').find('option:selected');
    var script = $(option).data('script');
    var env = $(option).data('env');

    if (!script || script.length === 0) {
        paramsElems.empty();
    } else {
        $.get('/ajax/script/params', {
            'script': script,
            'environment': env
        }, function (params) {
            paramsElems.empty();
            $.each(params, function () {
                paramsElems.append('<div class="col">' +
                                   '  <input type="text" class="form-control" id="' + this + '"name="' + this + '" placeholder="' + this + '" required/>' +
                                   '</div>');
            });
            paramsElems.append('<input type="hidden" name="environment" value="' + env + '"/>');
        });
    }
}

function updateEventHistory() {
    var eventLogContainer = $('.event-log');
    $.get('/ajax/events', function (events) {
        if (!events) {
            return;
        }
        eventLogContainer.empty();
        for (var mac in events) {
            var title = mac;
            if (events[mac][0].host != '') {
                var host = events[mac][0].server.Hostname;
                if (host == '') host = events[mac][0].server.IP;
                title += ' (' + host + ')';
            }
            var elem = eventLogContainer.append('<div class="card" id="' + mac + '"><h5 class="card-header text-primary-custom">' + title + '</h5><div class="card-body"><ul class="list-group list-group-flush">');

            $.each(events[mac], function () {
                var date = (new Date(this.date)).toLocaleString();
                var params = "";
                for (var p in this.params) {
                    params += p + ':' + this.params[p] + ' ';
                }
                elem.append('<li class="list-group-item"><b>' + date + '</b>: ' + this.message + '</li>');
            });

            eventLogContainer.append('</ul></div></div>');
        }
    });
}
