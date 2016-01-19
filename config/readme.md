## To be done

    Add zookeeper etc. support with an adapter interface.

## ini格式说明

相关库: github.com/go-ini/ini

    1. 区块继承使用.隔开,例如[production.testing],父类必须写在前面.
    2. 支持其他变量代换: %(work.dir)s
    3. 字符串不需要加引号.
