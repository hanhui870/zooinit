# this allow code file use utf-8 code chinese
# coding=utf-8
import os
import traceback


def test():
    try:
        a = u'bats\u00E0'
        print(a)
        print(a.encode("utf8"))
        # UnicodeEncodeError: 'ascii' codec can't encode character u'\xe0' in position 4: ordinal not in range(128)
        print(str(a))
    except Exception as err:
        print(err)
        # 中国String
        print(traceback.format_exc())
        # traceback.print_exc()


if __name__ == "__main__":
    test()
