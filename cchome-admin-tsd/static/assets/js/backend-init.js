define(['backend'], function (Backend) {
    //设置页面title
    var navTitle = $("#firstnav li.active span", window.parent.document);
    if (navTitle.html() == "") {
        navTitle.html(document.title);   
    }
    if (Config.addtabs.sider_url != "") {
        $("ul.sidebar-menu li.active", window.parent.document).removeClass("active");
        $($("ul.sidebar-menu li a[url='" + Config.addtabs.sider_url + "']", window.parent.document).parents("li")).addClass("active");
        $("ul.sidebar-menu li.active ul", window.parent.document).removeAttr("style");
        $($("ul.sidebar-menu li a[url='" + Config.addtabs.sider_url + "']", window.parent.document).parents("ul")[0]).addClass("menu-open").attr("style", "display:block");

        //顶栏的icon
        $("#firstnav li.active a i", window.parent.document).removeAttr("class").addClass(Config.addtabs.icon);
        
        // var nav = $("#firstnav li.active", window.parent.document)
        // nav.attr("id","tab_"+document.title);
        // $("a",nav).attr("href","#con_"+document.title).attr("node-id",document.title).attr("aria-controls",document.title);


        // console.log(document.);
        //存储到localstorage
        // <a href="/station/info?id=561040037651677185" url="/station/info?id=561040037651677185" addtabs="航盛科技大厦-站点详情" class="hide"><i class="fa fa-university"></i> <span>航盛科技大厦-站点详情</span></a>
        // localStorage.setItem("addtabs",'<a href="'+document.URL+'" url="/station/info?id=561040037651677185" addtabs="航盛科技大厦-站点详情" class="hide"><i class="fa fa-university"></i> <span>航盛科技大厦-站点详情</span></a>')
    }
});