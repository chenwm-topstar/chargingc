define(['jquery', 'bootstrap', 'backend', 'table', 'form'], function ($, undefined, Backend, Table, Form) {
    var Controller = {
        search: function () {
            if (self.frameElement && self.frameElement.tagName == "IFRAME") {
                parent.location.reload();
                return;
            }

            Form.api.bindevent($("form[role=form]"), function(data, ret){
                setTimeout(function () {
                    console.log(data);
                    window.location = data['url'];
                }, 1000);
            });
        },
    };
    return Controller;
});