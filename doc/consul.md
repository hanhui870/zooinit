 # Consul startup

 ## Server startup command

    consul agent -server -data-dir="/tmp/consul" -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108

