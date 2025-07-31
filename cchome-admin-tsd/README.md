# init




# 2.0.1
- 上线步骤
- 1. 添加设备绑定用户的appid
    - update evse_bind eb JOIN users u on (eb.uid=u.id) set eb.appid = u.appid;
- 2. 洗数据，将设备的is_master 赋值
    - update evse_bind eb LEFT JOIN (select min(id) as id from evse_bind group by appid, sn) tmp on (eb.id = tmp.id)
      set eb.is_master=1 where tmp.id is not null;
- 3. 开启配置

    {"send_host":"smtp.exmail.qq.com","send_port":465,"user_name":"info@goiot.net","passwd":"yIP25qQ7","def_sender_mail":"info@goiot.net"}

