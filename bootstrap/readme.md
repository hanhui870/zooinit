# Service bootstrap

Bootstrop the basic etcd based high available discovery service for low level use.

## Usage

zooinit bootstrap -f config/config.ini

## Synopsis

    1. First bootstrap etcd service of discovery in configuraion file. Then register local service self in to registry.
    2. Second boot other etcd servers in the intranet.
    3. Finally the bootstrap service is up when qurorum reach qurorum size configured in the file.