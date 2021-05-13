#!/usr/bin/env python2
import numpy as np




def map(name, contents):
    #dataset = contents.split()

    kva = []
    test_dataset = []
    part_train_dataset = []
    tmp_dataset = contents.split("\n")
    for i in tmp_dataset:
        if len(i) == 0:
            continue
        info = i.split(" ")
        p_c = int(float(info[1]))

        if info[0] == "train":
            p = np.array([float(j) for j in info[2:]])
            part_train_dataset.append((p_c, p))
        else:

            p = np.array([float(j) for j in info[3:]])
            test_dataset.append((p_c, p))
    k = 5
    for (q_c, q) in test_dataset:
        lisdict = {}
        dist_list = []
        class_list = []

        for (p_c, p) in part_train_dataset:
            dis = np.sum(np.square(p - q))
            dist_list.append(dis)
            class_list.append(p_c)
        arg_label = np.argsort(np.array(dist_list))[0:k]
        written_in = np.array(dist_list)[arg_label]
        written_in_label =np.array(class_list)[arg_label]
        input_string = ''
        for z in range(k):
            input_string += (str(written_in[z]) + ':' + str(written_in_label[z]) + ' ')
	key_string = str(q_c)
        lisdict[key_string] = input_string[:-1]
        kva.append(lisdict)

    return kva


def reduce(key, values):

    k = 5
    distance_list = []
    label_list = []

    for listdict in values:
        input_item = listdict
        written_in = input_item.split(' ')
        for q in written_in:
            data = q.split(':')
            distance, label = float(data[0]), data[1]
            distance_list.append(distance)
            label_list.append(int(label))
    result = np.argsort(distance_list)[0:k]
    result = np.array(label_list)[result]

    return np.argmax(np.bincount(result))
