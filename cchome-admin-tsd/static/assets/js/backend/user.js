define(['jquery', 'bootstrap', 'backend', 'table', 'form', 'validator'], function ($, undefined, Backend, Table, Form) {

    var Controller = {
        list: function () {
            // 初始化表格参数配置
            Table.api.init({
                extend: {
                    index_url: 'user/list',
                }
            });

            var table = $("#table");

            //在表格内容渲染完成后回调的事件
            table.on('post-body.bs.table', function (e, json) {
                $("tbody tr[data-index]", this).each(function (idx, v) {
                });
            });
            // 初始化表格
            table.bootstrapTable({
                method: "post",
                url: $.fn.bootstrapTable.defaults.extend.index_url,
                columns: [
                    [
                        {field: 'id', title: "user id"},
                        {field: 'app_client', title: 'app name'},
                        {field: 'name', title: 'user name'},
                        {field: 'email', title: "email"},
                    ]
                ],
            });

            // 为表格绑定事件
            Table.api.bindevent(table);
        },
        logoff: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
    };
    return Controller;
});



