#! /usr/bin/python3

import sys
import string


def str_to_unicode_hex(s: str) -> list[str]:
    newStr = []
    for c in s:
        if not str.isspace(c):
            c = "\\u{:0>4x}".format(ord(c))

        newStr.append(c)

    return ''.join(newStr)

def convert_file(name: str):
    with open(name) as f:
        converted = [str_to_unicode_hex(line) for line in f]
        return ''.join(converted)

def ascii_table(width= 8):
    start, end = ord('!'), ord('~')
    str_list = []
    for x in range(start, end+1):
        hexval = f'{x:0>4X}'
        s = f" {hexval}|{chr(x)} "
        str_list.append(s)


    print("ASII Table")
    while len(str_list) > 0:
        until = min(width, len(str_list))
        print("".join(str_list[0:until]))
        del str_list[0:until]

if __name__ == '__main__':
    args = sys.argv[1:]
    if not args or args[0] == '-h' or args[0] == 'help':
        print("Use it like this: ")
        print("\t- unicode_escape_generator.py table")
        print("\t- unicode_escape_generator.py [filename]")
    elif args[0] == 'table':
        ascii_table()
    else:
        print(convert_file(args[0]))
    
