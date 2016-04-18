# Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
import signal
import os
import sys
import time


def signalHandler(signum, frame):
    print('Signal handler called by:', signum, " will exit now.")
    sys.exit(1)


def registerExitSignal():
    signal.signal(signal.SIGINT, signalHandler)
    # signal.signal(signal.SIGKILL, signalHandler) # not support
    signal.signal(signal.SIGTERM, signalHandler)
    signal.signal(signal.SIGABRT, signalHandler)
    signal.signal(signal.SIGSEGV, signalHandler)


if __name__ == "__main__":
    print("Test system signal")
    registerExitSignal()
    time.sleep(60)
