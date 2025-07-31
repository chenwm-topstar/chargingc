define(['jquery', 'bootstrap', 'backend', 'table', 'form', 'validator'], function ($, undefined, Backend, Table, Form) {

    var Controller = {
        list: function () {
            // 初始化表格参数配置
            Table.api.init({
                    extend: {
                    index_url: 'qaa/list',
                    add_url: 'qaa/add',
                    edit_url: 'qaa/edit',
                    del_url: 'qaa/del',
                }
            });

            $(".btn-add").data("shade", 0.5);
            $(".btn-add").data("area", ["95%","95%"]);
            $(".btn-add").data("title", "添加参数配置");

            var table = $("#table");

            //在表格内容渲染完成后回调的事件
            table.on('post-body.bs.table', function (e, json) {
                $("tbody tr[data-index]", this).each(function (idx, v) {
                    $(".btn-editone").data("shade", 0.5);
                    $(".btn-editone").data("area", ["95%","95%"]);
                    $(".btn-editone").data("title", "change Q&A");
                });
            });
            // 初始化表格
            table.bootstrapTable({
                method: "post",
                url: $.fn.bootstrapTable.defaults.extend.index_url,
                columns: [
                    [
                        {field: 'id', title: "ID"},
                        {field: 'q', title: "Question"},
                        {field: 'a', title: "Answer"},
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