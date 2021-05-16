# Golang implement of MapReduce

This is our EE447 final project, idea comes from MIT 6.824 course project. Contributors are @[**sun-lingyu**](https://github.com/sun-lingyu), @[**yifanlu0227**](https://github.com/yifanlu0227),@[**Nicholas0228**](https://github.com/Nicholas0228)



### Required package & how to install

- golang 1.15+

- crypto/ssh : go get golang.org/x/crypto/ssh@v0.0.0-20201221181555-eec23a3978ad

- python3-dev : sudo apt-get install python3-dev -y



### Optional package & how to install

- nltk (for word count & inverted index example ): pip install nltk -y
- numpy (for KNN example): pip install numpy



### Usage

First run ` git clone https://github.com/yifanlu0227/mapreduce.git ` to download this resposity to your machines. Select one machine to be the coordinator, and  others to be workers. 

You should edit your worker's ip / username / password in `mapreduce/src/main/mrcoordinator.go` like following. 

```go
	hosts := []string{"192.168.0.132", "192.168.0.184", "192.168.0.33", "192.168.0.199"}
	command := "go run mrworker.go " + os.Args[1]
	mr.AwakenWorkers("root", "Ydhlw123", hosts, command)
```

And you should make sure the 1234 port and 8081 port are available, since we will use them for our RPC and http server.



### Python support

Our MapReduce support python development, i.e., you can just provide a simple python file including *map* function and *reduce* function. You can refer to our provide example like word count `mapreduce/src/main/wc.py` .

```python
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
```

To run the this word count example with input file `pg-*.txt` ,  run this in terminal

```go
go run mrcoordinator.go wc pg-*.txt
```

The KNN example

```go
go run mrcoordinator.go knn dataset*.txt
```

The Inverted Index example

```go
go run mrcoordinator.go inverted_index pg-*.txt
```

To see the output file, run

```sh
cat mr-out-* | sort | more
```



### Experiment

**word count**:

<img src="image/wordcount.png" alt="wordcount" style="zoom:83%;" />

---

**inverted index**:

<img src="image/invertedindex.png" alt="invertedindex" style="zoom:83%;" />

---

**KNN large dataset:**

<img src="image/KNN.png" alt="wordcount" style="zoom:83%;" />

### Visualization

**worker perspective**

<img src="image/worker.jpg" alt="worker" style="zoom:50%;" />

---

**file perspective**

<img src="image/file.jpg" alt="file" style="zoom:50%;" />

<img src="image/file2.jpg" alt="file" style="zoom:50%;" />

### Acknowledge

MIT 6.824
