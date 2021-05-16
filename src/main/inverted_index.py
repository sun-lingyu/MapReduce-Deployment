#!/usr/bin/env python2
import sys

import os

import string
import nltk


def map(name, contents):
	lower = contents.upper()
	remove =  string.maketrans(string.punctuation, string.punctuation,) 
	lower1 = lower.translate(remove, string.punctuation,)
	without_punctuation = lower1.translate(remove, string.digits,)
	tokens = nltk.word_tokenize(without_punctuation)
	kva = []
	for p in tokens:
		lisdict = {}
		lisdict[p] = name
		kva.append(lisdict)
	return kva
def reduce(key, values):
	inputlist = sorted(values)
	count = 0
	output = ""
	for i in range(len(inputlist) - 1):
		count += 1
		if inputlist[i] != inputlist[i+1]:
			output = output + " " + inputlist[i] + " " +str(count)
			count = 0
	count += 1
	output = output + " " + inputlist[-1] + " " +str(count)
	return output


