import subprocess
import sys
import io
from cluster.info import Info


def run(info):
    if (not isinstance(info, Info)):
        print(__name__ + "::run() info is not instance Info, please check")

    try:
        # fail, no quoted: consul agent -server -data-dir="/tmp/consul" -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108
        # If passing a single string, either shell must be True (see below) or else the string must simply name the program to be executed without specifying any arguments.
        proc = subprocess.Popen(["consul", "agent", "-server",
                                 "-data-dir=/tmp/consul",
                                 "-bootstrap-expect", info.Qurorum,
                                 "-bind=" + info.Localip,
                                 "-client=" + info.Localip], stdout=subprocess.PIPE)

        for line in iter(proc.stdout.readline, ''):
            print(line.strip())

            # No need.
            # proc.wait()

    except subprocess.CalledProcessError as err:
        print("Found CalledProcessError:", err, err.output)
    except Exception as err:
        print("Found error:", err)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    run('test')
