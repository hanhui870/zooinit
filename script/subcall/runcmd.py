import subprocess
import sys


def runWithStdoutSync(args):
    try:
        # fail, no quoted: consul agent -server -data-dir="/tmp/consul" -bootstrap-expect 3  -bind=192.168.4.108 -client=192.168.4.108
        # If passing a single string, either shell must be True (see below) or else the string must simply name the program to be executed without specifying any arguments.
        proc = subprocess.Popen(args, stdout=subprocess.PIPE, bufsize=1, universal_newlines=True)

        with proc.stdout as out:
            while True:
                line = out.readline()
                if line != "":
                    print(line.strip())
                else:
                    break

                    # No need.
                    # proc.wait()

    except subprocess.CalledProcessError as err:
        print("Found CalledProcessError:", err, err.output)
        print("Will exit now...")
        sys.exit(1)

    except Exception as err:
        print("Found error:", err)
        print("Will exit now...")
        sys.exit(1)


# ImportError: No module named cluster.utils
# see readme.md set PYTHONPATH
if __name__ == "__main__":
    runWithStdoutSync(["date"])
    runWithStdoutSync(["ls", "-l"])
    runWithStdoutSync(["consul", "agent", "-server"])

    # will exit
    # runWithStdoutSync(["fdfdss"])

    runWithStdoutSync(["consul", "agent", "-server",
                       "-data-dir=/tmp/consul",
                       "-bootstrap-expect", "3",
                       "-bind=192.168.4.108",
                       "-client=192.168.4.108"])
