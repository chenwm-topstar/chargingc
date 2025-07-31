/*
 * Copyright (c) 2018.
 */

define(['jquery', 'bootstrap', 'backend', 'table', 'form'], function ($, undefined, Backend, Table, Form) {
    var Controller = {
        index: function () {
            if (self.frameElement && self.frameElement.tagName == "IFRAME") {
                parent.location.reload();
                return;
            }

            Form.api.bindevent($("form[role=form]"), function(data, ret){
                setTimeout(function () {
                    window.location = data['url'];
                }, 1000);
            });
        },
    };
    return Controller;
});