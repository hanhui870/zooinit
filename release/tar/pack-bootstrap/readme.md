# Mac start



# Useful command

    bin/zooinit boot --log.channel=file --boot.cmd=bin/etcd --log.path=`pwd`/log &

    crontab: cd /home/bruce/software/zooinit-1.0.0 && bin/zooinit boot --log.channel=file --boot.cmd=bin/etcd --log.path=`pwd`/log &