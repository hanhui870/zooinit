# Zoopy

## Abount PYTHONPATH

    Need to set export PYTHONPATH=`pwd`"/script" through shell.

```
export PYTHONPATH=`pwd`"/script"
```

## Python3 depends

1. http.client
1. https://github.com/jplana/python-etcd.git
1. ...


## Build python3 library

1. docker run -ti -v /Users/bruce/project/godev/src/zooinit/script/library/:/usr/local/python haimi:python3-dev
1. git clone https://github.com/jplana/python-etcd.git
1. python3 setup.py install --install-lib=$PYTHONPATH