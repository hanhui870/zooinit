# 分布式服务自动助手

主要解决分布式服务的启动问题:
1. 分布式服务的配置问题
1. 分布式服务的服务发现问题
1. 集群启动支持脚本调用和变量自动复制,基于Go语言模板解析

## Testing

    go test zooinit/log zooinit/discovery zooinit/config -v