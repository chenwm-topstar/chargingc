# 编译流程
- 1. 安装golang研发环境，版本: v1.19
- 2. 在当前目录执行命令: GOOS=linux GOARCH=amd64 go build -o app
- 3. 编译成功后, 可以看到当前目录下会生成一个app文件

# 打包流程
- 1. 需要自己准备docker仓库
- 2. 执行命令: docker build -t 仓库路径:版本 . 
- 3. docker push 仓库路径:版本

# 发布流程
- 1. 需要先熟悉k8s部署系统
- 2. 修改/root/k8s/cchome-admin/deployment.yaml 中的image路径
- 3. 执行命令 kubctrl apply -f /root/k8s/cchome-admin/deployment.yaml 


