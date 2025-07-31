define(['jquery', 'bootstrap', 'backend', 'table', 'form', 'validator'], function ($, undefined, Backend, Table, Form) {

    var Controller = {
        list: function () {
            // 初始化表格参数配置
            Table.api.init({
                extend: {
                    index_url: 'feedback/list',
                    edit_url: 'feedback/edit'
                    // del_url: 'feedback/del',
                }
            });

  
            var table = $("#table");
            // 初始化表格
            table.bootstrapTable({
                method: "post",
                url: $.fn.bootstrapTable.defaults.extend.index_url,
                columns: [
                    [
                        {field: 'id', title: "ID"},
                        {field: 'user_name', title: "user name"},
                        {field: 'user_email', title: "user email"},
                        {field: 'content', title: "feedback content"},
                        {field: 'is_process', title: "is process", formatter: function (value, row, index) {
                            return value
                        }},
                        {field: 'remark', title: "remark"},
                        {field: 'updated_at', title: "update time"},
                        {field: 'operate', title: "option", table: table, events: Table.api.events.operate, formatter: function (value, row, index) {
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