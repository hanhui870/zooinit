import sys


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
