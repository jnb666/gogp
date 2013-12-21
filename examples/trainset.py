#!/usr/bin/python
# generate data set for input to symbolic regression example
from math import *

def get_num(prompt):
    while True:
        try:
            return float(raw_input(prompt))
        except ValueError:
            print "invalid numeric value"

def get_func(prompt):
    while True:
        try:
            return eval("lambda x: " + raw_input(prompt))
        except SyntaxError:
            print "invalid function"

fun   = get_func("enter function where x is input variable: ")
start = get_num("enter start of x range: ")
end   = get_num("enter end of x range: ")
step  = get_num("enter step: ")
erc_start = get_num("enter start of random constant range: ")
erc_end   = get_num("enter end of random constant range: ")

filename = raw_input("enter output file: ")

with open(filename, 'w') as f:
    f.write("{0}\t{1}\n".format(erc_start, erc_end))
    x = start
    while x <= end:
        f.write("{0}\t{1}\n".format(x, fun(x)))
        x += step


