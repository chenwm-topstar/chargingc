define(['jquery', 'bootstrap', 'backend', 'table', 'form', 'validator'], function ($, undefined, Backend, Table, Form) {

    var Controller = {
        startcharger: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
        stopcharger: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
        reboot: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
        latestversion: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
        upgrade: function() {    
            $('input#upgrade_ftp').bind('keyup', function () {
                var filter = $('input#upgrade_ftp').val();
              
                $.ajax({
                    url: "/device/firmware/search",
                    type: "GET",
                    dataType: 'json',
                    data: {'filter': filter},
                    async: false,
                    success: function (res) {
                        if (res.data != null && res.data.urls != null) {
                            $('datalist#upgrade_list').empty();
                            for (var i = 0; i < res.data.urls.length; i++) {
                                var add_options = '<option value="' + res.data.urls[i] + '">'+ res.data.urls[i] + '</option>';
                                $('datalist#upgrade_list').append(add_options);
                            }
                        }
                    }
                })
            });

            Form.api.bindevent($("form[role=form]"));  
        },
        setconfig: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
        getappointconfig: function() {
            Form.api.bindevent($("form[role=form]"),function(data, ret){
                //如果我们需要在提交表单成功后做跳转，可以在此使用location.href="链接";进行跳转
                // Toastr.success("成功");
                $("#config_key").val(data["params"])
                return false
            });            
        },
        selectenv: function() {
            Form.api.bindevent($("form[role=form]"));            
        },
        list: function () {
            // 初始化表格参数配置
            Table.api.init({
                extend: {
                    index_url: 'device/list',
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
                        {field: 'id', title: "device id",},
                        {field: 'pn', title: 'model',formatter:function(value, row, index){
                            return "<a href='/device/info?evseid=" + row["id"] + "' class='btn addtabsit' data-shade='0.5' data-area='[\"98%\",\"98%\"]' title='" + "设备详情'>" + value + "</a>"
                        }},
                        {field: 'sn', title: 'sn',formatter:function(value, row, index){
                            return "<a href='/device/info?evseid=" + row["id"] + "' class='btn addtabsit' data-shade='0.5' data-area='[\"98%\",\"98%\"]' title='" + "设备详情'>" + value + "</a>"
                        }},
                        {field: 'state', title: "state",},
                        {field: 'firmware_version', title: "firmware version"},
                        {field: 'protocol_version', title: "protocol version"},
                        {field: 'rated_current', title: "rated current"},
                        {field: 'mac', title: "mac address"},
                    ]
                ],
            });

            // 为表格绑定事件
            Table.api.bindevent(table);
        },
        info: function () {
            $("ul.nav-tabs li[aria-name='"+Config.evse.id+":"+Config.evse.no+"']").addClass("active");
            
            Table.api.init({
                method: "post",
            });
           
            Controller.table['evse'].call(this);
            Controller.table['connector'].call(this);
        },
        table: {
            evse: function () {
                var evsetab = $("#evse");
                evsetab.bootstrapTable({
                    url: 'device/info?evseid='+Config.evse.evseid+"&no="+Config.evse.no,
                    showHeader: false,
                    columns: [
                        [
                            {field: 'key', title:'',align: 'right'}, 
                            {field: 'val', title: '',align: 'left'},                           
                        ]
                    ]
                });
                setInterval(function () {
                    evsetab.bootstrapTable('refresh', {silent: true});
                },30000);
                Table.api.bindevent(evsetab);
            },
            connector: function () {
                var table = $("#connector");
                table.bootstrapTable({
                    method: "post",
                    url:  'device/connector?evseid='+Config.evse.evseid,
                    columns: [
                        [
                            { field: 'id', title: "ID" },
                            { field: 'cno', title:"no"},
                            { field: 'state', title:"state", width:"100px", operate:false,formatter: function (value, row, index) {
                                    var desc = ""
                                    var custom = {
                                        "Unavailable" :"red",
                                        "Idle" :"success",
                                        "Connected" :"info",
                                        "Charging" :"yellow",
                                        "Charging pile is not output" :"blue",
                                        "Electric vehicle not charged" :"blue",
                                        "Charge completed" :"lime",
                                        "In Reserver" :"purple",
                                        "Failure" :"fuchsia",
                                        "Waiting" :"primary",
                                        "Occupying": "primary",
                                    };
                                    switch (value){
                                        case 0:  desc = "Unavailable"; break;
                                        case 1:  desc = "Idle"; break;
                                        case 2:  desc = "Connected"; break;
                                        case 3:  desc = "Charging"; break;
                                        case 4:  desc = "Charging pile is not output"; break;
                                        case 5:  desc = "Electric vehicle not charged"; break;
                                        case 6:  desc = "Charge completed"; break;
                                        case 7:  desc = "In Reserver"; break;
                                        case 8:  desc = "Failure"; break;
                                        case 9:  desc = "Waiting"; break;
                                        case 10: desc = "Occupying";break;
                                    }

                                    if (typeof this.custom !== 'undefined') {
                                        custom = $.extend(custom, this.custom);
                                    }
                                    this.icon = 'fa fa-circle';
                                    this.custom = custom;
                                    return Table.api.formatter.normal.call(this, desc, row, index);
                                }},
                            { field: 'power', title: 'Power(KW)'},
                            { field: 'current', title:"Current(A)"},
                            { field: 'voltage', title:"Voltage(V)"},
                            { field: 'electricity', title:"electric(KW.H)"},
                            { field: 'record_id', title: "record_id",formatter: function (value, row, index) {
                                if (value == "" ){
                                    return "-"
                                }
                                return value;
                            }},
                            { field: 'fault_code', title:"fault code"},
                            {
                                field: 'actions',
                                title: "command",
                                table: table,
                                events: Table.api.events.operate,
                                buttons: [
                                    {
                                        title: "tagget telemetry",
                                        classname: 'btn btn-xs btn-success btn-ajax',
                                        icon: 'fa fa-repeat',
                                        url: '/device/evse/triggerconnectorstate?&evseid='+Config.evse.evseid,
                                        extend:'style="width:25px"'
                                    },
                                    {
                                        title: "start charge",
                                        classname: 'btn btn-xs btn-success btn-ajax',
                                        icon: 'fa fa-flash',
                                        url: '/device/evse/start_charger?&evseid='+Config.evse.evseid,
                                        extend:'style="width:25px"'
                                    },
                                    {
                                        title: "stop charge",
                                        classname: 'btn btn-xs btn-danger btn-ajax',
                                        icon: 'fa fa-ban',
                                        url: '/device/evse/stop_charger?evseid='+Config.evse.evseid,
                                        extend:'style="width:25px"',
                                        success: function (data, ret) {
                                            var alertindex = Layer.alert(ret.msg, function(result) {
                                                table.bootstrapTable('refresh',{silent: true});
                                                Layer.close(alertindex);
                                            });
                                        }
                                    },
                                ],
                                formatter: Table.api.formatter.buttons
                            },
                        ]
                    ],
                });
            
                setInterval(function () {
                    table.bootstrapTable('refresh', {silent: true});
                },30000);
           
                // 为表格绑定事件
                Table.api.bindevent(table);
            }
        },
    };
    return Controller;
});



