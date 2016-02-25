import sys


# A unbuffered stdout, still has bug, total line by line may has to use rpc logger
class Unbuffered(object):
    def __init__(self, stream):
        self.stream = stream

    def write(self, data):
        self.stream.write(data)
        self.stream.flush()

    def __getattr__(self, attr):
        return getattr(self.stream, attr)


def initUnbufferedStdoutIO():
    sys.stdout = Unbuffered(sys.stdout)


# print and flush
def printf(*args, **kwargs):
    for var in args:
        print(var)
    for var in kwargs:
        print(var + "=" + str(kwargs[var]))

    # flush output will log in a new line.
    sys.stdout.flush()


if __name__ == "__main__":
    printf("test1", 2, test=222)
    initUnbufferedStdoutIO()
    print("test output.")
