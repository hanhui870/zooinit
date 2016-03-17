# coding=utf-8
# this allow code file use utf-8 code chinese
# python -c 'import locale; print( locale.getpreferredencoding())'
# reveal bug in test code. UnicodeEncodeError: 'ascii' codec can't encode character
import os
import subprocess
import traceback


def test():
    try:
        a = u'bats\u00E0\xb5'
        print(a)
        print(a.encode("utf8"))

        args = ["echo", a]
        proc = subprocess.Popen(args, stdout=subprocess.PIPE, bufsize=1)

        with proc.stdout as out:
            while True:
                line = out.readline()
                if line != b"":
                    line = line.strip().decode("utf8")
                    if line != "":
                        print(line.encode("utf8").decode("utf8"))
                else:
                    # print("End of stdout, will break out loop...")
                    break

        proc.wait()

        # UnicodeEncodeError: 'ascii' codec can't encode character u'\xe0' in position 4: ordinal not in range(128)
        # print(str(a))
    except Exception as err:
        print(err)
        # String
        print(traceback.format_exc())
        # traceback.print_exc()


if __name__ == "__main__":
    test()
