# 分布式服务自动助手

主要解决分布式服务的启动问题:
1. 分布式服务的配置问题
1. 分布式服务的服务发现问题
1. 集群启动支持脚本调用和变量自动复制,基于Go语言模板解析


    Note: software configuration watch and reoconfigure please use consul-template project

## Testing

    go test zooinit/log zooinit/discovery zooinit/config -v


## Roadmap

    1. TODO: Lately need to trigger reconfig cluster size
    2. TODO: Support service name command specified apart from backend name. 2016.03.14 done
    3. TODO: ELK and etc centralized logger facility support.
    4. TODO: ip.hint can pass by cmd args.
