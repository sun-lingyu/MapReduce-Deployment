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
		lisdict[p] = "1"
		kva.append(lisdict)
	return kva
	
def reduce(key, values):
	return str(len(values))

'''
result = {}
dictf = {}

for q in os.listdir( "../main/"):
	if q[0] != 'p':
		continue
	if q[1] != 'g':
		continue
	filepath = os.path.join("../main/", q)
	f = open(filepath)
	contents= f.read()
	output = map("a", contents)

	for p in output:
		index = hash(list(p.keys())[0])
		dictf[index] = list(p.keys())[0]
		if result.__contains__(index):
			result[index] += 1
		else:
			result[index] = 1

final_result= {}
for q,v in result.items():
	final_result[dictf[q]] = v

items = final_result.items() 
items.sort()
count = 0
for z in items:
	count = count+1
	if count > 50:
		break
	print z
'''

