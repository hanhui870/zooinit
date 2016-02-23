import sys
import os
from cluster.consul import start


# this style need full path cluster.consul.start.hello()
# import cluster.consul.start

# sys.path include pwd
def main():
    start.run()


if __name__ == "__main__":
    main()
