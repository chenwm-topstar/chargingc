define(['jquery', 'bootstrap', 'backend', 'table', 'form', 'validator'], function ($, undefined, Backend, Table, Form) {

    var Controller = {
        edit: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
    };
    return Controller;
});