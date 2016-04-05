# Zoopy

## Abount PYTHONPATH

    Need to set export PYTHONPATH=`pwd`"/script" through shell.

```
export PYTHONPATH=`pwd`"/script"
```

## Abount Specific qurorum usage

    Cluster script can handle qurorum specific usage for its own purpose.

## Python3 depends

1. http.client
1. https://github.com/jplana/python-etcd.git
1. ...


## Build python3 library

PYTHONPATH defined in .bash_profile

1. docker run -ti -v $PYTHONPATH:/usr/local/python haimi:python3-dev
1. git clone https://github.com/jplana/python-etcd.git
1. python3 setup.py install --install-lib=$PYTHONPATH

## pip install

1. pip3 install dnspython3 urllib3  -t $PYTHONPATH