## To be done

    Add zookeeper etc. support with an adapter interface.

## ini格式说明

相关库: github.com/go-ini/ini

    1. 区块继承使用.隔开,例如[production.testing],父类必须写在前面.
    2. 支持其他变量代换: %(work.dir)s 变量引用配置不当会导致panic, 有继承情况下可正常获取.
    3. 字符串不需要加引号.
    4. 父子继承的配置不能相同名称, 不然会导致递归调用.
