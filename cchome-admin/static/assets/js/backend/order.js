define(['jquery', 'bootstrap', 'backend', 'table', 'form', 'validator'], function ($, undefined, Backend, Table, Form) {

    var Controller = {
        list: function () {
            // 初始化表格参数配置
            Table.api.init({
                extend: {
                    index_url: 'order/list',
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
                        {field: 'id', title: "order id"},
                        {field: 'uid', title: 'uid'},
                        {field: 'record_id', title: "record_id"},
                        {field: 'sn', title: "SN"},
                        {field: 'auth_id', title: 'auth_id'},
                        {field: 'auth_mode', title: "auth_mode"},
                        {field: 'start_time', title: "start_time"},
                        {field: 'charge_time', title: "charge_time"},
                        {field: 'total_electricity', title: "total_electricity"},
                        {field: 'stop_reason', title: "stop_reason"},
                    ]
                ],
            });

            // 为表格绑定事件
            Table.api.bindevent(table);
        },
    };
    return Controller;
});



