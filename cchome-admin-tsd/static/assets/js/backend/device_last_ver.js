define(['jquery', 'bootstrap', 'backend', 'table', 'form', 'validator'], function ($, undefined, Backend, Table, Form) {
    var Controller = {
        list: function () {
            // 初始化表格参数配置
            Table.api.init({
                extend: {
                    index_url: 'dlv/list',
                    add_url: 'dlv/add',
                    edit_url: 'dlv/edit',
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
                        {field: 'id', title: "id",},
                        {field: 'pn', title: 'model'},
                        {field: 'vendor', title: "vendor",},
                        {field: 'last_version', title: "firmware version"},
                        {field: 'upgrade_address', title: "protocol version"},
                        {field: 'updated_at', title: "update time"},
                        {field: 'created_at', title: "create time"},
                        {field: 'operate', title: "cmd", table: table, events: Table.api.events.operate, formatter: function (value, row, index) {
                            return Table.api.formatter.operate.call(this, value, row, index);
                        }}
                    ]
                ],
            });

            // 为表格绑定事件
            Table.api.bindevent(table);
        },
        add: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
        edit: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
    };
    return Controller;
});



